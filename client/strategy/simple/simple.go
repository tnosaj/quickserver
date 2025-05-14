package simple

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/tnosaj/quickserver/client/internals"
)

type SimpleStrategy struct {
	S   internals.Settings
	Max int
}

func MakeSimpleStrategy(s internals.Settings) *SimpleStrategy {
	logrus.Info("Simple strategy")
	return &SimpleStrategy{S: s, Max: 0}
}

func (st *SimpleStrategy) RunCommand() {
	logrus.Debugf("ping %s", st.S.Url)
	perc := int(rand.Intn(100))
	switch {
	case perc <= 10:
		st.curl("DELETE")
	case perc <= 40:
		st.curl("POST")
	default:
		st.curl("GET")
	}
	logrus.Debugf("pong %s", st.S.Url)
}

func (st *SimpleStrategy) curl(method string) {
	timer := prometheus.NewTimer(st.S.Metrics.RequestDuration.WithLabelValues(method))

	client := http.Client{}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(st.S.Timeout)*time.Second)
	defer cancel()

	if method == "POST" || st.Max == 0 {
		st.Max += 1
	}
	i := int(rand.Intn(st.Max))
	logrus.Tracef("Will perform %s for %d", method, i)
	url := fmt.Sprintf("%s/entry/%d", st.S.Url, i)
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		logrus.Errorf("error creating request %s", err)
		return
	}

	res, err := client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	// res, err := http.Get(st.S.Url)

	if e, ok := err.(net.Error); ok && e.Timeout() {
		st.S.Metrics.ErrorRequests.WithLabelValues("timeout").Inc()
		logrus.Errorf("Error %s url timeout: %s - %s", method, url, err)
	} else if err == nil && res.StatusCode != http.StatusOK {
		st.S.Metrics.ErrorRequests.WithLabelValues(strconv.Itoa(res.StatusCode)).Inc()
		if res.StatusCode != http.StatusNotFound {
			logrus.Errorf("Error %s url reponsecode: %s - %d", method, url, res.StatusCode)
		}
	} else if err != nil {
		st.S.Metrics.ErrorRequests.WithLabelValues("unknown").Inc()
		logrus.Errorf("Error from an unknown source: %s", err)
	}
	timer.ObserveDuration()
}
