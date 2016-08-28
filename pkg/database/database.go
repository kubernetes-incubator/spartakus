package database

import (
	"fmt"

	"k8s.io/spartakus/pkg/logr"
	"k8s.io/spartakus/pkg/report"
)

type Database interface {
	Store(report.Record) error
}

func NewDatabase(log logr.Logger, dbspec string) (Database, error) {
	for name, plug := range plugins {
		is, db, err := plug.Attempt(log, dbspec)
		if !is {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create %q database: %v", name, err)
		}
		if db == nil {
			return nil, fmt.Errorf("%q database was nil", name)
		}
		log.V(0).Infof("using %q database", name)
		return db, nil
	}

	return nil, fmt.Errorf("unknown database: %q", dbspec)
}