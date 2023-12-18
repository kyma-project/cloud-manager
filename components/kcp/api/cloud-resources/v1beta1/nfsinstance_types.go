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
)

// +kubebuilder:validation:Enum=Regional;Zonal
type AwsFileSystemType string

const (
	AwsFileSystemTypeRegional = "Regional"
	AwsFileSystemTypeZonal    = "Zonal"
)

// +kubebuilder:validation:Enum=Enhanced;Bursting
type AwsThroughputMode string

const (
	AwsThroughputModeEnhanced = "Enhanced"
	AwsThroughputModeBursting = "Bursting"
)

// NfsInstanceSpec defines the desired state of NfsInstance
type NfsInstanceSpec struct {
	// +kubebuilder:validation:Required
	Kyma string `json:"kyma"`

	// +kubebuilder:validation:Required
	Scope *ScopeRef `json:"scope"`

	// +optional
	Gcp *NfsInstanceGcp `json:"gcp,omitempty"`

	// +optional
	Azure *NfsInstanceAzure `json:"azure,omitempty"`

	// +optional
	Aws *NfsInstanceAws `json:"aws,omitempty"`
}

type NfsInstanceGcp struct {
}

type NfsInstanceAzure struct {
}

type NfsInstanceAws struct {
	Type       AwsFileSystemType `json:"type,omitempty"`
	Throughput AwsThroughputMode `json:"throughput,omitempty"`
}

// NfsInstanceStatus defines the observed state of NfsInstance
type NfsInstanceStatus struct {
	State StatusState `json:"state,omitempty"`

	// List of status conditions to indicate the status of a Peering.
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NfsInstance is the Schema for the nfsinstances API
type NfsInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NfsInstanceSpec   `json:"spec,omitempty"`
	Status NfsInstanceStatus `json:"status,omitempty"`
}

func (in *NfsInstance) KymaName() string {
	return in.Spec.Kyma
}

func (in *NfsInstance) ScopeRef() *ScopeRef {
	return in.Spec.Scope
}

func (in *NfsInstance) SetScopeRef(scopeRef *ScopeRef) {
	in.Spec.Scope = scopeRef
}

func (in *NfsInstance) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
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
