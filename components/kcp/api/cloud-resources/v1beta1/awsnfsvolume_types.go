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
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of AwsNfsVolume. Edit awsnfsvolume_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// AwsNfsVolumeStatus defines the observed state of AwsNfsVolume
type AwsNfsVolumeStatus struct {
	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// AwsNfsVolume is the Schema for the awsnfsvolumes API
type AwsNfsVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsNfsVolumeSpec   `json:"spec,omitempty"`
	Status AwsNfsVolumeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// AwsNfsVolumeList contains a list of AwsNfsVolume
type AwsNfsVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsNfsVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsNfsVolume{}, &AwsNfsVolumeList{})
}
