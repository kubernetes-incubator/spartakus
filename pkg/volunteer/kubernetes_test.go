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
	kresource "k8s.io/client-go/1.4/pkg/api/resource"
	kv1 "k8s.io/client-go/1.4/pkg/api/v1"
)

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
			expect: report.Node{
				CloudProvider: strPtr("unknown"),
			},
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
				CloudProvider: strPtr("unknown"),
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
				CloudProvider:           strPtr("unknown"),
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
