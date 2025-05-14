package internals

import "github.com/prometheus/client_golang/prometheus"

type Settings struct {
	Debug   bool
	Port    string
	Timeout int

	Concurrency int
	Duration    int
	Rate        int

	DurationType string
	Strategy     string

	Url string

	Metrics Metrics
}

// Metrics contsins all metric types
type Metrics struct {
	RequestDuration *prometheus.HistogramVec
	ErrorRequests   *prometheus.CounterVec
}
