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
	"github.com/elliotchance/pie/v2"
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GcpNfsVolumeState string

// Valid GcpNfsVolume States.
const (
	// GcpNfsVolumeReady signifies GcpNfsVolume is completed and is Ready for use.
	GcpNfsVolumeReady GcpNfsVolumeState = "Ready"

	// GcpNfsVolumeError signifies the GcpNfsVolume operation resulted in error.
	GcpNfsVolumeError GcpNfsVolumeState = "Error"

	// GcpNfsVolumeProcessing signifies nfs GcpNfsVolume operation is in-progress.
	GcpNfsVolumeProcessing GcpNfsVolumeState = "Processing"

	// GcpNfsVolumeCreating signifies nfs create operation is in-progress.
	GcpNfsVolumeCreating GcpNfsVolumeState = "Creating"

	// GcpNfsVolumeUpdating signifies nfs update operation is in-progress.
	GcpNfsVolumeUpdating GcpNfsVolumeState = "Updating"

	// GcpNfsVolumeDeleting signifies nfs delete operation is in-progress.
	GcpNfsVolumeDeleting GcpNfsVolumeState = "Deleting"
)

// Additional error reasons
const (
	ConditionReasonCapacityInvalid         = "CapacityGbInvalid"
	ConditionReasonIpRangeNotReady         = "IpRangeNotReady"
	ConditionReasonFileShareNameInvalid    = "FileShareNameInvalid"
	ConditionReasonTierInvalid             = "TierInvalid"
	ConditionReasonPVNotReadyForDeletion   = "PVNotReadyForDeletion"
	ConditionReasonPVNotReadyForNameChange = "PVNotReadyForNameChange"
	ConditionReasonPVNameInvalid           = "PVNameInvalid"
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

	PersistentVolume *GcpNfsVolumePvSpec `json:"volume,omitempty"`

	PersistentVolumeClaim *GcpNfsVolumePvcSpec `json:"volumeClaim,omitempty"`
}

type GcpNfsVolumePvSpec struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type GcpNfsVolumePvcSpec struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
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

	//State of the GcpNfsVolume
	State GcpNfsVolumeState `json:"state,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Path",type="string",JSONPath=".spec.fileShareName"
// +kubebuilder:printcolumn:name="Capacity (GB)",type="string",JSONPath=".status.capacityGb"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

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

func (in *GcpNfsVolume) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfs
}

func (in *GcpNfsVolume) SpecificToProviders() []string {
	return []string{"gcp"}
}

//+kubebuilder:object:root=true

// GcpNfsVolumeList contains a list of GcpNfsVolume
type GcpNfsVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpNfsVolume `json:"items"`
}

func (l *GcpNfsVolumeList) GetItemCount() int {
	return len(l.Items)
}

func (l *GcpNfsVolumeList) GetItems() []client.Object {
	return pie.Map(l.Items, func(item GcpNfsVolume) client.Object {
		return &item
	})
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
