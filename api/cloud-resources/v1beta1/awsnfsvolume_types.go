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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

// AwsNfsVolumeSpec defines the desired state of AwsNfsVolume
type AwsNfsVolumeSpec struct {

	// +kubebuilder:validation:Required
	IpRange IpRangeRef `json:"ipRange"`

	// +kubebuilder:validation:Required
	Capacity resource.Quantity `json:"capacity"`

	// +kubebuilder:default=generalPurpose
	PerformanceMode AwsPerformanceMode `json:"performanceMode,omitempty"`

	// +kubebuilder:default=bursting
	Throughput AwsThroughputMode `json:"throughput,omitempty"`
}

// AwsNfsVolumeStatus defines the observed state of AwsNfsVolume
type AwsNfsVolumeStatus struct {

	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	Server string `json:"server,omitempty" json:"server,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=".status.conditions[?(@.type==\"Ready\")].status"
// +kubebuilder:printcolumn:name="Error",type="string",JSONPath=".status.conditions[?(@.type==\"Error\")].status"

// AwsNfsVolume is the Schema for the awsnfsvolumes API
type AwsNfsVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsNfsVolumeSpec   `json:"spec,omitempty"`
	Status AwsNfsVolumeStatus `json:"status,omitempty"`
}

func (in *AwsNfsVolume) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AwsNfsVolume) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

//+kubebuilder:object:root=true

// AwsNfsVolumeList contains a list of AwsNfsVolume
type AwsNfsVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsNfsVolume `json:"items"`
}

func (l *AwsNfsVolumeList) GetItemCount() int {
	return len(l.Items)
}

func init() {
	SchemeBuilder.Register(&AwsNfsVolume{}, &AwsNfsVolumeList{})
}
