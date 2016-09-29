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
	"strings"
	"testing"
	"time"

	"github.com/kubernetes-incubator/spartakus/pkg/database"
	"github.com/kubernetes-incubator/spartakus/pkg/report"
	"github.com/kubernetes-incubator/spartakus/pkg/version"
	logrtest "github.com/thockin/logr/testing"
)

// Fake out "list nodes" calls.
type fakeNodeLister struct {
	returnValue []report.Node
	returnError error
}

type fakeNamespaceLister struct {
	returnValue report.NamespaceStats
	returnError error
}

var _ nodeLister = fakeNodeLister{}

func (fake fakeNodeLister) ListNodes() ([]report.Node, error) {
	return fake.returnValue, fake.returnError
}

func (fake fakeNamespaceLister) ListNamespace() (report.NamespaceStats, error) {
	return fake.returnValue, fake.returnError
}

// Fake out "get server version" calls.
type fakeServerVersioner struct {
	returnValue string
	returnError error
}

var _ serverVersioner = fakeServerVersioner{}

func (fake fakeServerVersioner) ServerVersion() (string, error) {
	return fake.returnValue, fake.returnError
}

const fakeClusterID = "cluster"
const fakePeriod = time.Hour

func newTestVolunteer(t *testing.T) *volunteer {
	log := &logrtest.TestLogger{T: t}
	db := database.Database(nil)
	nodes := &fakeNodeLister{}
	vers := &fakeServerVersioner{}
	namespace := &fakeNamespaceLister{}
	return newVolunteer(log, fakeClusterID, fakePeriod, db, nodes, namespace, vers)
}

func TestGenerateRecord(t *testing.T) {
	testCases := []struct {
		tweak     func(vol *volunteer)
		errstr    string
		version   string
		nodes     []string
		namespace int
	}{
		{ // test serverVersioner failure
			tweak: func(vol *volunteer) {
				vol.serverVersioner.(*fakeServerVersioner).returnError = fmt.Errorf("fail")
			},
			errstr: "fail",
		},
		{ // test nodeLister failure
			tweak: func(vol *volunteer) {
				vol.nodeLister.(*fakeNodeLister).returnError = fmt.Errorf("fail")
			},
			errstr: "fail",
		},
		{ // test namespaceLister failure
			tweak: func(vol *volunteer) {
				vol.namespaceLister.(*fakeNamespaceLister).returnError = fmt.Errorf("fail")
			},
			errstr: "fail",
		},
		{ // test success
			tweak: func(vol *volunteer) {
				vol.serverVersioner.(*fakeServerVersioner).returnValue = "v1.2.3"
				vol.namespaceLister.(*fakeNamespaceLister).returnValue = report.NamespaceStats{
					Total: 15,
				}
				vol.nodeLister.(*fakeNodeLister).returnValue = []report.Node{
					{ID: "node1"}, {ID: "node2"},
				}
			},
			version:   "v1.2.3",
			nodes:     []string{"node1", "node2"},
			namespace: 15,
		},
	}

	for i, tc := range testCases {
		vol := newTestVolunteer(t)
		tc.tweak(vol)
		rec, err := vol.generateRecord()
		if err != nil && tc.errstr == "" {
			t.Errorf("[%d] unexpected error %q", i, err)
		} else if err != nil {
			if !strings.Contains(err.Error(), tc.errstr) {
				t.Errorf("[%d] expected error containing %q, got %q", i, tc.errstr, err)
			}
		} else if err == nil && tc.errstr != "" {
			t.Errorf("[%d] expected error containing %q: no error", i, tc.errstr)
		} else {
			if rec.ClusterID != fakeClusterID {
				t.Errorf("[%d] expected ClusterID %q, got %q", i, fakeClusterID, rec.ClusterID)
			}
			if rec.Version != version.VERSION {
				t.Errorf("[%d] expected Version %q, got %q", i, version.VERSION, rec.Version)
			}
			if *rec.MasterVersion != tc.version {
				t.Errorf("[%d] expected MasterVersion %q, got %q", i, rec.MasterVersion, tc.version)
			}
			if len(rec.Nodes) != len(tc.nodes) {
				t.Errorf("[%d] expected %d nodes, got %d", i, len(rec.Nodes), len(tc.nodes))
			}
			if rec.Namespace.Total != tc.namespace {
				t.Errorf("[%d] unexpected number of namespaces to be %v, got %v ", i, rec.Namespace.Total, tc.namespace)
			}
			for j := range rec.Nodes {
				if rec.Nodes[j].ID != tc.nodes[j] {
					t.Errorf("[%d] expected node[%d].ID %q, got %q", i, j, rec.Nodes[j].ID, tc.nodes[j])
				}
			}
		}
	}
}
