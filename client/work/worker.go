package work

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tnosaj/quickserver/client/internals"
	"github.com/tnosaj/quickserver/client/strategy"
	"github.com/tnosaj/quickserver/client/strategy/simple"
)

// ExecutionType defines how long queries are run
type ExecutionType interface {
	Run()
}

// Start running
func Start(s internals.Settings) {

	wp := newWorkerPool(s.Concurrency, s.Rate)

	var st strategy.ExecutionStrategy

	switch s.Strategy {
	case "simple":
		st = simple.MakeSimpleStrategy(s)
	}

	go run(s, wp, st)
}

// WorkerPool is a struct with a channel and workgroup
type workerPool struct {
	ch      chan func()
	wg      *sync.WaitGroup
	limiter <-chan time.Time
}

// NewWorkerPool creates a pool of size N
func newWorkerPool(poolsize, rate int) *workerPool {
	interval := rate2interval(rate)
	logrus.Infof("Rate limit is %d", interval)
	limiter := time.Tick(interval * time.Millisecond)
	wg := &sync.WaitGroup{}
	ch := make(chan func())
	for i := 0; i < poolsize; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()
			for f := range ch {
				f()
			}
		}()
	}
	return &workerPool{ch: ch, wg: wg, limiter: limiter}
}

// Do - workerpool DOES
func (w *workerPool) do(f func()) {
	// if duration is zero, time.Tick returns nil
	if w.limiter != nil {
		<-w.limiter
	}
	w.ch <- f
}

// Stop - workerpool STOPES
func (w *workerPool) stop() {
	close(w.ch)
	w.wg.Wait()
}

func rate2interval(rate int) time.Duration {
	// milliseconds / rate per second
	if rate > 0 {
		return time.Duration(1000 / rate)
	}
	return time.Duration(0)
}
