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

package report

type Record struct {
	// Version is the version.VERSION of the schema being reported.
	Version string `json:"version"` // required
	// Timestamp is the UNIX timestamp when the report was received.
	Timestamp string `json:"timestamp"` // provided by server, client values are ignored
	// ClusterID is a string reported by the volunteer.  It could be anything but
	// a random GUID is strongly recommended. This should be a stable value for
	// the lifetime of the cluster, or else reports will not be assumed to be
	// the same cluster.  This must not include personally identifiable
	// information.
	ClusterID string `json:"clusterID"` // required
	// MasterVersion is the version string of the kubernetes master in the
	// reporting cluster.
	MasterVersion *string `json:"masterVersion,omitempty"`
	// Nodes is a list of node-specific information from the reporting cluster.
	Nodes []Node `json:"nodes,omitempty"`
}

type Node struct {
	// ID is a unique string that identifies a node in tis cluster.  It can be
	// any value but we strongly recommend a random GUID or a hash derived from
	// identifying information.  This should be a stable value for the lifetime
	// of the node, or else it will be assumed to be a different node.  This
	// must not include personally identifiable information.
	ID string `json:"id"` // required
	// OperatingSystem is the value reported by kubernetes in the node status.
	OperatingSystem *string `json:"operatingSystem,omitempty"`
	// OSImage is the value reported by kubernetes in the node status.
	OSImage *string `json:"osImage,omitempty"`
	// KernelVersion is the value reported by kubernetes in the node status.
	KernelVersion *string `json:"kernelVersion,omitempty"`
	// Architecture is the value reported by kubernetes in the node status.
	Architecture *string `json:"architecture,omitempty"`
	// ContainerRuntimeVersion is the value reported by kubernetes in the node
	// status.
	ContainerRuntimeVersion *string `json:"containerRuntimeVersion,omitempty"`
	// KubeletVersion is the value reported by kubernetes in the node status.
	KubeletVersion *string `json:"kubeletVersion,omitempty"`
	// CloudProvider is the <ProviderName> portion of the ProviderID reported
	// by kubernetes in the node spec.
	CloudProvider *string `json:"cloudProvider,omitempty"`
	// Capacity is a list of resources and their associated values as reported
	// by kubernetes in the node status.
	Capacity []Resource `json:"capacity,omitempty"`
}

type Resource struct {
	// Resource is the name of the resource.
	Resource string `json:"resource"` // required
	// Value is the string form of the of the resource's value.
	Value string `json:"value"` // required
}
