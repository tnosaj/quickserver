package main

import (
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/tnosaj/quickserver/client/internals/server"
)

func main() {
	s, err := evaluateInputs()
	if err != nil {
		log.Fatalf("could not evaluate inputs: %q", err)
	}
	setupLogger(s.Debug)

	// Generate seedyness
	rand.New(rand.NewSource(time.Now().UnixNano()))

	server := server.NewHttpbenchServer(s)

	router := mux.NewRouter()
	router.Handle("/metrics", promhttp.Handler())
	router.HandleFunc("/status", server.Status).Methods("GET")
	router.HandleFunc("/run", server.Run).Methods("POST")
	log.Fatal(http.ListenAndServe(":"+s.Port, router))
}

func setupLogger(debug bool) {
	//logrus.SetReportCaller(true)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	logrus.Debug("Configured logger")
}
