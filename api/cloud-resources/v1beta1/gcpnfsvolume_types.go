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
	"k8s.io/apimachinery/pkg/api/resource"
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
	ConditionReasonPVCNameInvalid          = "PVCNameInvalid"
	ConditionReasonNoWorkerZones           = "NoWorkerZones"
	ConditionReasonLocationInvalid         = "LocationInvalid"
	ConditionReasonTierLegacy              = "LegacyTier"
)

// GcpNfsVolumeSpec defines the desired state of GcpNfsVolume
// +kubebuilder:validation:XValidation:rule=(self.tier != "REGIONAL" || self.tier == "REGIONAL" && (self.capacityGb >= 1024 && self.capacityGb <= 9984 && (self.capacityGb - 1024) % 256 == 0 || self.capacityGb >= 10240 && self.capacityGb <= 102400 && (self.capacityGb - 10240) % 2560 == 0)), message="REGIONAL tier capacityGb must be between 1024 and 9984, and it must be divisble by 256, or between 10240 and 102400, and divisible by 2560"
// +kubebuilder:validation:XValidation:rule=(self.tier != "REGIONAL" || self.tier == "REGIONAL" && size(self.fileShareName) <= 64), message="REGIONAL tier fileShareName length must be 64 or less characters"
// +kubebuilder:validation:XValidation:rule=(self.tier != "ZONAL" || self.tier == "ZONAL" && (self.capacityGb >= 1024 && self.capacityGb <= 9984 && (self.capacityGb - 1024) % 256 == 0 || self.capacityGb >= 10240 && self.capacityGb <= 102400 && (self.capacityGb - 10240) % 2560 == 0)), message="ZONAL tier capacityGb must be between 1024 and 9984, and it must be divisble by 256, or between 10240 and 102400, and divisible by 2560"
// +kubebuilder:validation:XValidation:rule=(self.tier != "ZONAL" || self.tier == "ZONAL" && size(self.fileShareName) <= 64), message="ZONAL tier fileShareName length must be 64 or less characters"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_SSD" || self.tier == "BASIC_SSD" && self.capacityGb >= 2560 && self.capacityGb <= 65400), message="BASIC_SSD tier capacityGb must be between 2560 and 65400"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_SSD" || self.tier == "BASIC_SSD" && size(self.fileShareName) <= 16), message="BASIC_SSD tier fileShareName length must be 16 or less characters"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_SSD" || self.tier == "BASIC_SSD" && self.capacityGb >= oldSelf.capacityGb), message="BASIC_SSD tier capacityGb cannot be reduced"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_HDD" || self.tier == "BASIC_HDD" && self.capacityGb >= 1024 && self.capacityGb <= 65400), message="BASIC_HDD tier capacityGb must be between 1024 and 65400"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_HDD" || self.tier == "BASIC_HDD" && size(self.fileShareName) <= 16), message="BASIC_HDD tier fileShareName length must be 16 or less characters"
// +kubebuilder:validation:XValidation:rule=(self.tier != "BASIC_HDD" || self.tier == "BASIC_HDD" && self.capacityGb >= oldSelf.capacityGb), message="BASIC_HDD tier capacityGb cannot be reduced"
type GcpNfsVolumeSpec struct {
	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="IpRange is immutable."
	IpRange IpRangeRef `json:"ipRange"`
	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Location is immutable."
	Location string `json:"location"`

	// +kubebuilder:default=BASIC_HDD
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Tier is immutable."
	Tier GcpFileTier `json:"tier"`

	// +kubebuilder:validation:Pattern="^[a-z][a-z0-9_]*[a-z0-9]$"
	// +kubebuilder:default=vol1
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="FileShareName is immutable."
	FileShareName string `json:"fileShareName"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="SourceBackup is immutable."
	SourceBackup GcpNfsVolumeBackupRef `json:"sourceBackup,omitempty"`

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
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Name is immutable."
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
	// Deprecated
	// +optional
	CapacityGb int `json:"capacityGb"`

	// Provisioned capacity
	// +optional
	Capacity resource.Quantity `json:"capacity"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	//State of the GcpNfsVolume
	State GcpNfsVolumeState `json:"state,omitempty"`

	Location string `json:"location,omitempty"`

	// +optional
	Protocol string `json:"protocol,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="Path",type="string",JSONPath=".spec.fileShareName"
// +kubebuilder:printcolumn:name="Capacity",type="string",JSONPath=".status.capacity"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// GcpNfsVolume is the Schema for the gcpnfsvolumes API
type GcpNfsVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpNfsVolumeSpec   `json:"spec,omitempty"`
	Status GcpNfsVolumeStatus `json:"status,omitempty"`
}

func (in *GcpNfsVolume) State() string {
	return string(in.Status.State)
}

func (in *GcpNfsVolume) SetState(v string) {
	in.Status.State = GcpNfsVolumeState(v)
}

func (in *GcpNfsVolume) GetIpRangeRef() IpRangeRef {
	return in.Spec.IpRange
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

// +kubebuilder:validation:Enum=BASIC_HDD;BASIC_SSD;ZONAL;REGIONAL
type GcpFileTier string

const (
	BASIC_HDD = GcpFileTier("BASIC_HDD")
	BASIC_SSD = GcpFileTier("BASIC_SSD")
	ZONAL     = GcpFileTier("ZONAL")
	REGIONAL  = GcpFileTier("REGIONAL")
)

func (in *GcpNfsVolume) CloneForPatchStatus() client.Object {
	return &GcpNfsVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GcpNfsVolume",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}
