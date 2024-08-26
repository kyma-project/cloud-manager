/*
Copyright 2023.

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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	ReasonFailedCreatingFileSystem        = "FailedCreatingFileSystem"
	ReasonInvalidMountTargetsAlreadyExist = "InvalidMountTargetsAlreadyExist"
)

// +kubebuilder:validation:Enum=generalPurpose;maxIO
type AwsPerformanceMode string

const (
	AwsPerformanceModeGeneralPurpose = AwsPerformanceMode("generalPurpose")
	AwsPerformanceModeBursting       = AwsPerformanceMode("maxIO")
)

// +kubebuilder:validation:Enum=bursting;elastic
type AwsThroughputMode string

const (
	AwsThroughputModeBursting = AwsThroughputMode("bursting")
	AwsThroughputModeElastic  = AwsThroughputMode("elastic")
)

// NfsInstanceSpec defines the desired state of NfsInstance
// +kubebuilder:validation:XValidation:rule=(has(self.instance.openStack) || false) && self.ipRange.name == "" || (has(self.instance.aws) || has(self.instance.gcp) || false) && self.ipRange.name != "", message="IpRange can not be specified for openstack, and is mandatory for gcp and aws."
type NfsInstanceSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="RemoteRef is immutable."
	RemoteRef RemoteRef `json:"remoteRef"`

	// +optional
	IpRange IpRangeRef `json:"ipRange"`

	// +kubebuilder:validation:Required
	Scope ScopeRef `json:"scope"`

	// +kubebuilder:validation:Required
	Instance NfsInstanceInfo `json:"instance"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type NfsInstanceInfo struct {
	// +optional
	Gcp *NfsInstanceGcp `json:"gcp,omitempty"`

	// +optional
	Azure *NfsInstanceAzure `json:"azure,omitempty"`

	// +optional
	Aws *NfsInstanceAws `json:"aws,omitempty"`

	// +optional
	OpenStack *NfsInstanceOpenStack `json:"openStack,omitempty"`
}

type NfsInstanceGcp NfsOptionsGcp

type NfsInstanceAzure struct {
}

type NfsInstanceOpenStack struct {
	// +kubebuilder:validation:Required
	SizeGb int `json:"sizeGb"`
}

type NfsInstanceAws struct {
	// +kubebuilder:default=generalPurpose
	PerformanceMode AwsPerformanceMode `json:"performanceMode,omitempty"`

	// +kubebuilder:default=bursting
	Throughput AwsThroughputMode `json:"throughput,omitempty"`
}

// NfsInstanceStatus defines the observed state of NfsInstance
type NfsInstanceStatus struct {
	State StatusState `json:"state,omitempty"`

	// +optional
	Id string `json:"id,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	//List of NFS Hosts (DNS Names or IP Addresses) that clients can use to connect
	//
	// XDeprecated: Use Host and Path
	// +optional
	Hosts []string `json:"hosts,omitempty"`

	// +optional
	Host string `json:"host,omitempty"`

	// +optional
	Path string `json:"path,omitempty"`

	// Operation Identifier to track the Hyperscaler Operation
	// +optional
	OpIdentifier string `json:"opIdentifier,omitempty"`

	// Provisioned Capacity in GBs
	// +optional
	CapacityGb int `json:"capacityGb"`

	// +optional
	StateData map[string]string `json:"stateData,omitempty"`
}

var _ client.Object = &NfsInstance{}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NfsInstance is the Schema for the nfsinstances API
// +kubebuilder:printcolumn:name="Scope",type="string",JSONPath=".spec.scope.name"
// +kubebuilder:printcolumn:name="SizeGb",type="string",JSONPath=".status.capacityGb"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"
type NfsInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NfsInstanceSpec   `json:"spec,omitempty"`
	Status NfsInstanceStatus `json:"status,omitempty"`
}

func (in *NfsInstance) ScopeRef() ScopeRef {
	return in.Spec.Scope
}

func (in *NfsInstance) SetScopeRef(scopeRef ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *NfsInstance) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *NfsInstance) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *NfsInstance) SetStateData(key, value string) {
	if in.Status.StateData == nil {
		in.Status.StateData = make(map[string]string)
	}
	if value == "" {
		delete(in.Status.StateData, key)
	} else {
		in.Status.StateData[key] = value
	}
}

func (in *NfsInstance) GetStateData(key string) (string, bool) {
	v, found := in.Status.StateData[key]
	return v, found
}

func (in *NfsInstance) CloneForPatchStatus() client.Object {
	return &NfsInstance{
		TypeMeta: metav1.TypeMeta{
			Kind:       "NfsInstance",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}

//+kubebuilder:object:root=true

// NfsInstanceList contains a list of NfsInstance
type NfsInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NfsInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NfsInstance{}, &NfsInstanceList{})
}
