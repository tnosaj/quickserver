package work

import (
	"github.com/sirupsen/logrus"
	"github.com/tnosaj/quickserver/client/internals"
	"github.com/tnosaj/quickserver/client/strategy"
)

func run(s internals.Settings, wp *workerPool, st strategy.ExecutionStrategy) {
	logrus.Infof("Running with a %s strategy", s.Strategy)

	// Catch other strategies

	var runner ExecutionType
	switch s.DurationType {
	case "events":
		runner = RunForEventCount{
			s:  s,
			wp: wp,
			st: st,
		}
	case "seconds":
		logrus.Fatalf("Sorry, seconds is not implemented yet")
	}

	runner.Run()
}

// RunForEventCount do stuffs
type RunForEventCount struct {
	s  internals.Settings
	wp *workerPool
	st strategy.ExecutionStrategy
}

// Run for a number of events
func (r RunForEventCount) Run() {

	for i := 0; i < r.s.Duration; i++ {
		r.wp.do(func() {
			r.st.RunCommand()
		})
	}
	r.wp.stop()

}
