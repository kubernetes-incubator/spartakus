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
	"reflect"
	"testing"

	"github.com/kubernetes-incubator/spartakus/pkg/report"
	"github.com/kylelemons/godebug/pretty"
	kclient "k8s.io/client-go/1.4/kubernetes"
	kresource "k8s.io/client-go/1.4/pkg/api/resource"
	kv1 "k8s.io/client-go/1.4/pkg/api/v1"
)

/*
 ok we need we have two inputs and they need to be faked.
*/

type kubeClientWrapperForTesting struct {
	client *kclient.Clientset
}

func TestNamespaceFromKubeNamespaces(t *testing.T) {
	testCase := struct {
		input  map[int][]float64
		except report.NamespaceStats
	}{
		input: map[int][]float64{
			Namespaces:  []float64{24},
			Pods:        []float64{5, 8, 1, 78},
			Deployments: []float64{7, 5, 453, 32},
			Jobs:        []float64{0, 2, 0, 0},
			CreationTS:  []float64{0.3, 12, 52},
			Services:    []float64{4, 0, 10, 0, 9},
		},
		except: report.NamespaceStats{Total: 24, PodsAvg: 23, DeployAvg: 124.3, JobsAvg: 0.5, LifetimeAvg: 21.4, ServiceAvg: 4.6},
	}

	ns := namespaceFromKubeNamespaces(testCase.input)
	if !reflect.DeepEqual(ns, testCase.except) {
		t.Errorf("[%d]: did not get expected result:\n", pretty.Compare(ns, testCase.except))
	}
}

func TestNodeFromKubeNode(t *testing.T) {
	testCases := []struct {
		input  kv1.Node
		expect report.Node
	}{
		{
			input: kv1.Node{
				ObjectMeta: kv1.ObjectMeta{
					Name: "kname",
				},
			},
			expect: report.Node{},
		},
		{
			input: kv1.Node{
				ObjectMeta: kv1.ObjectMeta{
					Name: "kname",
				},
				Status: kv1.NodeStatus{
					Capacity: kv1.ResourceList{
						// unsorted
						"r2": kresource.MustParse("200"),
						"r3": kresource.MustParse("300"),
						"r1": kresource.MustParse("100"),
					},
				},
			},
			expect: report.Node{
				Capacity: []report.Resource{
					{Resource: "r1", Value: "100"},
					{Resource: "r2", Value: "200"},
					{Resource: "r3", Value: "300"},
				},
			},
		},
		{
			input: kv1.Node{
				ObjectMeta: kv1.ObjectMeta{
					Name: "kname",
				},
				Status: kv1.NodeStatus{
					NodeInfo: kv1.NodeSystemInfo{
						OperatingSystem:         "os",
						OSImage:                 "image",
						KernelVersion:           "kernel",
						Architecture:            "architecture",
						ContainerRuntimeVersion: "runtime",
						KubeletVersion:          "kubelet",
					},
				},
			},
			expect: report.Node{
				OperatingSystem:         strPtr("os"),
				OSImage:                 strPtr("image"),
				KernelVersion:           strPtr("kernel"),
				Architecture:            strPtr("architecture"),
				ContainerRuntimeVersion: strPtr("runtime"),
				KubeletVersion:          strPtr("kubelet"),
			},
		},
	}

	for i, tc := range testCases {
		n := nodeFromKubeNode(&tc.input)
		if n.ID == "" || n.ID == tc.input.Name {
			t.Errorf("[%d] expected anonymized ID, got %q", i, n.ID)
		}
		n.ID = ""
		if !reflect.DeepEqual(n, tc.expect) {
			t.Errorf("[%d]: did not get expected result:\n%s", i, pretty.Compare(n, tc.expect))
		}
	}
}
