/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"github.com/kubernetes-incubator/spartakus/pkg/database"
	"github.com/kubernetes-incubator/spartakus/pkg/report"
	"github.com/kubernetes-incubator/spartakus/pkg/version"
	"github.com/thockin/logr"
)

var (
	CollectorEndpoint = "/api/v1"
	HealthEndpoint    = "/healthz"
	VersionEndpoint   = "/version"
)

type APIServer struct {
	Log      logr.Logger
	Port     int
	Database database.Database
}

func (s *APIServer) Run() error {
	handler := handlers.LoggingHandler(logWriter{s.Log}, s.newHandler())
	srv := &http.Server{
		Addr:    net.JoinHostPort("", strconv.Itoa(s.Port)),
		Handler: handler,
	}

	s.Log.V(0).Infof("binding to %s", srv.Addr)

	return srv.ListenAndServe()
}

// An adapter for gorilla's LoggingHandler.
type logWriter struct {
	logr.Logger
}

func (lw logWriter) Write(b []byte) (int, error) {
	lw.V(0).Infof(string(b))
	return len(b), nil
}

func (s *APIServer) newHandler() http.Handler {
	m := httprouter.New()
	m.Handle("GET", "/", s.healthHandler())
	m.Handle("POST", CollectorEndpoint, s.storeRecordHandler())
	m.Handle("GET", HealthEndpoint, s.healthHandler())
	m.Handle("GET", VersionEndpoint, s.versionHandler())
	return m
}

func (s *APIServer) storeRecordHandler() httprouter.Handle {
	handle := func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to read record: %v, err", err))
			return
		}

		var rec report.Record
		if err := json.Unmarshal(body, &rec); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("failed to decode record: %v", err))
			return
		}
		s.logRecord(&rec)

		if err := s.Database.Store(rec); err != nil {
			writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to store record: %v", err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
	return contentTypeMiddleware(handle, "application/json")
}

func (s *APIServer) logRecord(r *report.Record) {
	if s.Log.V(9).Enabled() {
		j, err := json.Marshal(r)
		if err != nil {
			s.Log.V(9).Infof("failed to decode record: %v", err)
		} else {
			s.Log.V(9).Infof("received record: %s", string(j))
		}
	}
}

func (s *APIServer) healthHandler() httprouter.Handle {
	return func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
	}
}

func (s *APIServer) versionHandler() httprouter.Handle {
	return func(w http.ResponseWriter, _ *http.Request, _ httprouter.Params) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(version.VERSION)); err != nil {
			s.Log.Errorf("failed writing version response: %v", err)
		}
	}
}
