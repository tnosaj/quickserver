package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/tnosaj/quickserver/client/internals"
)

func evaluateInputs() (internals.Settings, error) {
	var s internals.Settings

	flag.BoolVar(&s.Debug, "v", false, "Enable verbose debugging output")

	flag.StringVar(&s.Port, "p", "8080", "Starts server on this port")

	flag.IntVar(&s.Timeout, "t", 1, "Timeout in seconds for a backend answer")

	flag.IntVar(&s.Concurrency, "c", 1, "Concurrent number of threads to run")
	flag.IntVar(&s.Duration, "d", 1, "The number of events to process")
	flag.IntVar(&s.Rate, "r", 0, "requests per second - 0 to disable rate limiting")

	flag.StringVar(&s.DurationType, "duration", "events", "Duratation type - events|seconds")
	flag.StringVar(&s.Strategy, "strategy", "simple", "Strategy to use")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s: [flags] command [command args…]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	return s, nil
}
