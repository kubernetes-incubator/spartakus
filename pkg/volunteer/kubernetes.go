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
	"crypto/md5"
	"encoding/hex"
	"sort"

	"github.com/kubernetes-incubator/spartakus/pkg/report"
	stat "github.com/kubernetes-incubator/spartakus/pkg/statistics"
	kclient "k8s.io/client-go/1.4/kubernetes"
	kapi "k8s.io/client-go/1.4/pkg/api"
	"k8s.io/client-go/1.4/pkg/api/unversioned"
	kv1 "k8s.io/client-go/1.4/pkg/api/v1"
	kv1b "k8s.io/client-go/1.4/pkg/apis/extensions/v1beta1"
	krest "k8s.io/client-go/1.4/rest"
)

const RoundIt = 1

const (
	Pods = iota
	Services
	Jobs
	Deployments
	CreationTS
	Namespaces
)

type nodeLister interface {
	ListNodes() ([]report.Node, error)
}

type namespaceLister interface {
	ListNamespace() (report.NamespaceStats, error)
}

type serverVersioner interface {
	ServerVersion() (string, error)
}

func nodeFromKubeNode(kn *kv1.Node) report.Node {
	n := report.Node{
		ID:                      getID(kn),
		OperatingSystem:         strPtr(kn.Status.NodeInfo.OperatingSystem),
		OSImage:                 strPtr(kn.Status.NodeInfo.OSImage),
		KernelVersion:           strPtr(kn.Status.NodeInfo.KernelVersion),
		Architecture:            strPtr(kn.Status.NodeInfo.Architecture),
		ContainerRuntimeVersion: strPtr(kn.Status.NodeInfo.ContainerRuntimeVersion),
		KubeletVersion:          strPtr(kn.Status.NodeInfo.KubeletVersion),
	}
	// We want to iterate the resources in a deterministic order.
	keys := []string{}
	for k, _ := range kn.Status.Capacity {
		keys = append(keys, string(k))
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := kn.Status.Capacity[kv1.ResourceName(k)]
		n.Capacity = append(n.Capacity, report.Resource{
			Resource: string(k),
			Value:    v.String(),
		})
	}
	return n
}

func namespaceFromKubeNamespaces(m map[int][]float64) report.NamespaceStats {
	n := report.NamespaceStats{
		Total:       int(m[Namespaces][0]),
		LifetimeAvg: stat.RoundPlus(stat.Mean(m[CreationTS]), RoundIt),
		PodsAvg:     stat.RoundPlus(stat.Mean(m[Pods]), RoundIt),
		DeployAvg:   stat.RoundPlus(stat.Mean(m[Deployments]), RoundIt),
		ServiceAvg:  stat.RoundPlus(stat.Mean(m[Services]), RoundIt),
		JobsAvg:     stat.RoundPlus(stat.Mean(m[Jobs]), RoundIt),
	}
	return n
}

func getID(kn *kv1.Node) string {
	// We don't want to report the node's Name - that is PII.  The MachineID is
	// apparently not always populated and SystemUUID is ill-defined.  Let's
	// just hash them all together.  It should be stable, and this reduces risk
	// of PII leakage.
	return hashOf(kn.Name + kn.Status.NodeInfo.MachineID + kn.Status.NodeInfo.SystemUUID)
}

func hashOf(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil)[0:])
}

func strPtr(str string) *string {
	if str == "" {
		return nil
	}
	p := new(string)
	*p = str
	return p
}

func getNamespaceLifetime(then unversioned.Time) float64 {
	delta := unversioned.Now().Sub(then.Time)
	return delta.Hours()
}

func newKubeClientWrapper() (*kubeClientWrapper, error) {
	kubeConfig, err := krest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	kubeClient, err := kclient.NewForConfig(kubeConfig)
	if err != nil {
		return nil, err
	}
	return &kubeClientWrapper{client: kubeClient}, nil
}

type kubeClientWrapper struct {
	client *kclient.Clientset
}

func (k *kubeClientWrapper) ListNamespace() (report.NamespaceStats, error) {
	var ns report.NamespaceStats
	knsl, err := k.client.Core().Namespaces().List(kapi.ListOptions{})
	if err != nil {
		return ns, err
	}
	d, err := collectNamespaceData(k, knsl)
	if err != nil {
		return ns, err
	}
	return namespaceFromKubeNamespaces(d), nil
}

func collectNamespaceData(k *kubeClientWrapper, knsl *kv1.NamespaceList) (map[int][]float64, error) {
	l := len(knsl.Items)
	m := map[int][]float64{
		Namespaces:  make([]float64, l),
		Pods:        make([]float64, l),
		Services:    make([]float64, l),
		Deployments: make([]float64, l),
		Jobs:        make([]float64, l),
		CreationTS:  make([]float64, l),
	}
	m[Namespaces][0] = float64(l)
	for i := range knsl.Items {
		s := &knsl.Items[i]
		// get creation timestam of namspace per namespace
		m[CreationTS][i] = getNamespaceLifetime(s.CreationTimestamp)
		// get list of Pods per namespace
		pl, err := getPodsForNamespace(k, s.Name)
		if err != nil {
			return nil, err
		}
		m[Pods][i] = float64(len(pl))
		// get list of Service per namespace
		sl, err := getServicesForNamespace(k, s.Name)
		if err != nil {
			return nil, err
		}
		m[Services][i] = float64(len(sl))
		// get list of Jobs per namespace
		jl, err := getJobsForNamespace(k, s.Name)
		if err != nil {
			return nil, err
		}
		m[Jobs][i] = float64(len(jl))
		//get list of Deployments per namespace
		dl, err := getDeploymentsForNamespace(k, s.Name)
		if err != nil {
			return nil, err
		}
		m[Deployments][i] = float64(len(dl))
	}
	return m, nil
}

func getPodsForNamespace(k *kubeClientWrapper, ns string) ([]kv1.Pod, error) {
	pl, err := k.client.Core().Pods(ns).List(kapi.ListOptions{})
	if err != nil {
		return nil, err
	}
	return pl.Items, nil
}

func getServicesForNamespace(k *kubeClientWrapper, ns string) ([]kv1.Service, error) {
	sl, err := k.client.Core().Services(ns).List(kapi.ListOptions{})
	if err != nil {
		return nil, err
	}
	return sl.Items, err
}

func getJobsForNamespace(k *kubeClientWrapper, ns string) ([]kv1b.Job, error) {
	jl, err := k.client.Extensions().Jobs(ns).List(kapi.ListOptions{})
	if err != nil {
		return nil, err
	}
	return jl.Items, nil
}

func getDeploymentsForNamespace(k *kubeClientWrapper, ns string) ([]kv1b.Deployment, error) {
	dl, err := k.client.Extensions().Deployments(ns).List(kapi.ListOptions{})
	if err != nil {
		return nil, err
	}
	return dl.Items, nil
}

func (k *kubeClientWrapper) ListNodes() ([]report.Node, error) {
	knl, err := k.client.Core().Nodes().List(kapi.ListOptions{})
	if err != nil {
		return nil, err
	}
	nodes := make([]report.Node, len(knl.Items))
	for i := range knl.Items {
		kn := &knl.Items[i]
		nodes[i] = nodeFromKubeNode(kn)
	}
	return nodes, nil
}

func (k *kubeClientWrapper) ServerVersion() (string, error) {
	info, err := k.client.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return info.String(), nil
}
