package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	_ "github.com/lib/pq"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	conn *sql.DB
	//ctx context.Context

	appDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_duration_seconds",
		Help:    "Duration of operations",
		Buckets: prometheus.LinearBuckets(0.02, 0.02, 100),
	}, []string{"method"})

	errorReuests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "app_errors",
			Help: "The total number of failed requests",
		},
		[]string{"method"},
	)
)

func main() {
	//ctx := context.Background()
	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)

	// Connect to DB
	var err error
	psqlInfo := "sslmode=disable " + psqlInfoFromConnectionInfo()
	conn, err = sql.Open("postgres", psqlInfo)
	conn.SetMaxIdleConns(20)
	conn.SetMaxOpenConns(20)
	conn.SetConnMaxLifetime(360 * time.Second)
	if err != nil {
		log.Fatalf("Failed to ping database: %s", err)
	}
	if err := conn.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %s", err)
	} else {
		log.Default().Printf("Successfully connected to db")
	}

	defer conn.Close()
	defer shutdown()

	// Ensure table exists
	if _, err := conn.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS entries (
		id INTEGER PRIMARY KEY,
		last_seen TIMESTAMPTZ NOT NULL DEFAULT now()
	)`); err != nil {
		log.Fatalf("Unable to create table: %v", err)
	}

	// Prometheus metrics
	prometheus.MustRegister(appDuration, errorReuests)

	// logger := httplog.NewLogger("httplog-example", httplog.Options{
	// 	// JSON:             true,
	// 	LogLevel:         slog.LevelDebug,
	// 	Concise:          true,
	// 	RequestHeaders:   true,
	// 	MessageFieldName: "message",
	// 	QuietDownRoutes: []string{
	// 		"/",
	// 		"/ping",
	// 		"/status",
	// 		"/metrics",
	// 	},
	// 	QuietDownPeriod: 60 * time.Second,
	// 	// SourceFieldName: "source",
	// })

	r := chi.NewRouter()
	//r.Use(httplog.RequestLogger(logger))
	//r.Use(middleware.Heartbeat("/ping"))

	r.Post("/entry/{id}", handleCreate)
	r.Get("/entry/{id}", handleGet)
	r.Delete("/entry/{id}", handleDelete)
	r.Get("/status", handleStatus)
	r.Handle("/metrics", promhttp.Handler())

	port := os.Getenv("PORT")
	fmt.Printf("Listening on :%s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), r))
}

func psqlInfoFromConnectionInfo() string {
	if len(os.Getenv("DATABASE_URL")) > 0 {
		return os.Getenv("DATABASE_URL")
	}
	return fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("DATABASE_USER"),
		os.Getenv("DATABASE_PASS"),
		os.Getenv("DATABASE_HOST"),
		os.Getenv("DATABASE_PORT"),
		os.Getenv("DATABASE_NAME"),
	)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(appDuration.WithLabelValues("set"))
	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)

	defer func() { timer.ObserveDuration(); shutdown() }()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		log.Default().Printf("Error casting write: %s", err)
		errorReuests.WithLabelValues("setcast").Inc()
		return
	}
	_, err = conn.ExecContext(ctx, "INSERT INTO entries (id) VALUES ($1) ON CONFLICT (id) DO UPDATE SET last_seen = now()", id)
	if err != nil {
		http.Error(w, fmt.Sprintf("db error: %s", err), http.StatusInternalServerError)
		log.Default().Printf("Error writting: %s", err)
		errorReuests.WithLabelValues("set").Inc()
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleGet(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(appDuration.WithLabelValues("get"))
	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)

	defer func() { timer.ObserveDuration(); shutdown() }()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		log.Default().Printf("Error casting get: %s", err)
		errorReuests.WithLabelValues("getcast").Inc()
		return
	}

	var lastSeen time.Time

	err = conn.QueryRowContext(ctx, "SELECT last_seen FROM entries where id = $1", id).Scan(&lastSeen)

	if err != nil {
		if err == sql.ErrNoRows {
			http.NotFound(w, r)
			return
		} else if err.Error() != "sql: no rows in result set" {
			http.NotFound(w, r)
			return
		}
		http.Error(w, fmt.Sprintf("db error: %s", err), http.StatusInternalServerError)
		log.Default().Printf("Error getting: %s", err)
		errorReuests.WithLabelValues("get").Inc()
		return
	}

	go func(id int) {
		myctx, myshutdown := context.WithTimeout(context.Background(), 5*time.Second)
		defer myshutdown()
		// Get and update last_seen
		_, err := conn.QueryContext(myctx, `UPDATE entries
		SET last_seen = now()
		WHERE id = $1
	`, id)
		if err != nil {
			log.Default().Printf("Async update failed: %s", err)
			errorReuests.WithLabelValues("asyncget").Inc()
		}
	}(id)

	fmt.Fprintf(w, `{"id": %d, "last_seen": "%s"}`, id, lastSeen.Format(time.RFC3339))
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(appDuration.WithLabelValues("del"))
	ctx, shutdown := context.WithTimeout(context.Background(), 5*time.Second)

	defer func() { timer.ObserveDuration(); shutdown() }()

	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid ID", http.StatusBadRequest)
		log.Default().Printf("Error casting delete: %s", err)
		errorReuests.WithLabelValues("delcast").Inc()
		return
	}

	res, err := conn.ExecContext(ctx, `DELETE FROM entries WHERE id = $1`, id)
	if err != nil {
		http.Error(w, fmt.Sprintf("db error: %s", err), http.StatusInternalServerError)
		log.Default().Printf("Error deleting: %s", err)
		errorReuests.WithLabelValues("del").Inc()
		return
	}

	rows, err := res.RowsAffected()
	if err != nil {
		http.Error(w, fmt.Sprintf("db error: %s", err), http.StatusInternalServerError)
		log.Default().Printf("Error deleting: %s", err)
		errorReuests.WithLabelValues("del").Inc()
		return
	}
	if rows == 0 {
		http.NotFound(w, r)
		return
	}

	w.WriteHeader(http.StatusOK)
}
