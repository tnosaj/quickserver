package server

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/tnosaj/quickserver/client/internals"
	"github.com/tnosaj/quickserver/client/work"
)

func (s *HttpbenchServer) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	io.WriteString(w, `{"status": "ok"}`)
}

func (s *HttpbenchServer) Run(w http.ResponseWriter, r *http.Request) {
	logrus.Debugf("Run request")
	settings, err := s.unifySettings(r)
	if err != nil {
		returnError(w, err, http.StatusInternalServerError)
		return
	}
	go work.Start(settings)
}

func (s *HttpbenchServer) unifySettings(r *http.Request) (internals.Settings, error) {
	settings := s.Settings
	httpsettings, err := getSettingsFromPost(r)
	if err != nil {
		return s.Settings, err
	}

	if httpsettings.Concurrency > 0 {
		settings.Concurrency = httpsettings.Concurrency
	}
	if httpsettings.Duration > 0 {
		settings.Duration = httpsettings.Duration
	}
	if httpsettings.Rate > 0 {
		settings.Rate = httpsettings.Rate
	}
	if httpsettings.Url != "" {
		settings.Url = httpsettings.Url
	}
	logrus.Infof("Settings %+v", settings)
	return settings, nil
}

func getSettingsFromPost(r *http.Request) (HttpSettings, error) {
	untypedRequestBody, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return HttpSettings{}, fmt.Errorf("body read error: %q", err)
	}

	var typedRequestBody HttpSettings
	err = json.Unmarshal(untypedRequestBody, &typedRequestBody)
	if err != nil {
		return HttpSettings{}, fmt.Errorf("untypedRequestBody Unmarshal error: %q", err)
	}
	return typedRequestBody, nil
}

func returnError(w http.ResponseWriter, err error, httpCode int) {
	logrus.Error(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpCode)
	io.WriteString(w, fmt.Sprintf(`{"error": "%s"}`, err))
}
