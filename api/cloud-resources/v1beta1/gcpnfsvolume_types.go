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

// Additional error reasons
const (
	ConditionReasonCapacityInvalid       = "CapacityGbInvalid"
	ConditionReasonIpRangeNotReady       = "IpRangeNotReady"
	ConditionReasonFileShareNameInvalid  = "FileShareNameInvalid"
	ConditionReasonTierInvalid           = "TierInvalid"
	ConditionReasonPVNotReadyForDeletion = "PVNotReadyForDeletion"
)

// GcpNfsVolumeSpec defines the desired state of GcpNfsVolume
type GcpNfsVolumeSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="IpRange is immutable."
	IpRange IpRangeRef `json:"ipRange"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Location is immutable."
	Location string `json:"location"`

	// +kubebuilder:default=BASIC_HDD
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Tier is immutable."
	Tier GcpFileTier `json:"tier"`

	// +kubebuilder:validation:Pattern="^[a-z][a-z0-9_]*[a-z0-9]$"
	// +kubebuilder:default=vol1
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="FileShareName is immutable."
	FileShareName string `json:"fileShareName"`

	// +kubebuilder:default=2560
	CapacityGb int `json:"capacityGb"`
}

// GcpNfsVolumeStatus defines the observed state of GcpNfsVolume
type GcpNfsVolumeStatus struct {

	// +optional
	Id string `json:"id,omitempty"`

	//List of NFS Hosts (DNS Names or IP Addresses) that clients can use to connect
	// +optional
	Hosts []string `json:"hosts,omitempty"`

	// Capacity of the volume with Ready Condition
	// +optional
	CapacityGb int `json:"capacityGb"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GcpNfsVolume is the Schema for the gcpnfsvolumes API
type GcpNfsVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpNfsVolumeSpec   `json:"spec,omitempty"`
	Status GcpNfsVolumeStatus `json:"status,omitempty"`
}

func (in *GcpNfsVolume) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *GcpNfsVolume) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

//+kubebuilder:object:root=true

// GcpNfsVolumeList contains a list of GcpNfsVolume
type GcpNfsVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpNfsVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpNfsVolume{}, &GcpNfsVolumeList{})
}

type GcpFileTier string

const (
	STANDARD       = GcpFileTier("STANDARD")
	PREMIUM        = GcpFileTier("PREMIUM")
	BASIC_HDD      = GcpFileTier("BASIC_HDD")
	BASIC_SSD      = GcpFileTier("BASIC_SSD")
	HIGH_SCALE_SSD = GcpFileTier("HIGH_SCALE_SSD")
	ENTERPRISE     = GcpFileTier("ENTERPRISE")
	ZONAL          = GcpFileTier("ZONAL")
	REGIONAL       = GcpFileTier("REGIONAL")
)
