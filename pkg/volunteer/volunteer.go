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

package volunteer

import (
	"fmt"
	"strconv"
	"time"

	"github.com/kubernetes-incubator/spartakus/pkg/database"
	"github.com/kubernetes-incubator/spartakus/pkg/report"
	"github.com/kubernetes-incubator/spartakus/pkg/version"
	"github.com/thockin/logr"
)

func New(log logr.Logger, clusterID string, period time.Duration, db database.Database) (*volunteer, error) {
	kcw, err := newKubeClientWrapper()
	if err != nil {
		return nil, err
	}
	return newVolunteer(log, clusterID, period, db, kcw, kcw), nil
}

func newVolunteer(
	log logr.Logger,
	clusterID string,
	period time.Duration,
	db database.Database,
	nodeLister nodeLister,
	serverVersioner serverVersioner) *volunteer {

	return &volunteer{
		log:             log,
		clusterID:       clusterID,
		period:          period,
		database:        db,
		nodeLister:      nodeLister,
		serverVersioner: serverVersioner,
	}
}

type volunteer struct {
	clusterID       string
	period          time.Duration
	database        database.Database
	log             logr.Logger
	nodeLister      nodeLister
	serverVersioner serverVersioner
}

func (v *volunteer) Run() error {
	v.log.V(0).Infof("started volunteer")
	for {
		if err := v.runOnce(); err != nil {
			v.log.Errorf("%v", err)
		} else {
			v.log.V(0).Infof("report successfully sent")
		}
		if v.period == 0 {
			return nil
		}
		v.log.V(0).Infof("next attempt in %v", v.period)
		<-time.After(v.period)
	}
	// This can never be reached, and `go vet` complains if code is here.
}

func (v *volunteer) runOnce() error {
	rec, err := v.generateRecord()
	if err != nil {
		return fmt.Errorf("failed generating report: %v", err)
	}

	if err = v.send(rec); err != nil {
		return fmt.Errorf("failed sending report: %v", err)
	}

	return nil
}

func (v *volunteer) generateRecord() (report.Record, error) {
	svrVer, err := v.serverVersioner.ServerVersion()
	if err != nil {
		return report.Record{}, err
	}

	nodes, err := v.nodeLister.ListNodes()
	if err != nil {
		return report.Record{}, err
	}

	rec := report.Record{
		Version:       version.VERSION,
		Timestamp:     strconv.FormatInt(time.Now().Unix(), 10),
		ClusterID:     v.clusterID,
		MasterVersion: &svrVer,
		Nodes:         nodes,
	}

	return rec, nil
}

func (v *volunteer) send(rec report.Record) error {
	return v.database.Store(rec)
}
