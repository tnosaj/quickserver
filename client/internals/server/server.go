package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/tnosaj/quickserver/client/internals"
)

type HttpbenchServer struct {
	Settings internals.Settings
}

type HttpSettings struct {
	Strategy    string `json:"strategy"`
	Concurrency int    `json:"concurrency"`
	Duration    int    `json:"duration"`
	Rate        int    `json:"rate"`
	Url         string `json:"url"`
}

func NewHttpbenchServer(settings internals.Settings) HttpbenchServer {
	logrus.Info("starting server")
	RequestDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_duration_seconds",
		Help:    "Histogram for the runtime of a simple method function.",
		Buckets: prometheus.LinearBuckets(0.02, 0.02, 100),
	}, []string{"method"})

	ErrorReuests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "request_errors",
			Help: "The total number of failed requests",
		},
		[]string{"code"},
	)

	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(ErrorReuests)

	settings.Metrics = internals.Metrics{
		RequestDuration: RequestDuration,
		ErrorRequests:   ErrorReuests,
	}
	return HttpbenchServer{Settings: settings}
}
