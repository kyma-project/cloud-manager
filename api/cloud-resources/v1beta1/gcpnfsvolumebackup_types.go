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
	featuretypes "github.com/kyma-project/cloud-manager/pkg/feature/types"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GcpNfsBackupState string

// Valid GcpNfsBackup States.
const (
	// GcpNfsBackupReady signifies Backup is completed and is Ready for use.
	GcpNfsBackupReady GcpNfsBackupState = "Ready"

	// GcpNfsBackupError signifies the backup operation resulted in error.
	GcpNfsBackupError GcpNfsBackupState = "Error"

	// GcpNfsBackupCreating signifies backup create operation is in-progress.
	GcpNfsBackupCreating GcpNfsBackupState = "Creating"

	// GcpNfsBackupDeleting signifies backup delete operation is in-progress.
	GcpNfsBackupDeleting GcpNfsBackupState = "Deleting"

	// GcpNfsBackupDeleted signifies backup delete operation is complete.
	GcpNfsBackupDeleted GcpNfsBackupState = "Deleted"

	// GcpNfsBackupFailed signifies backup operation failed, and it will not be retried again.
	GcpNfsBackupFailed GcpNfsBackupState = "Failed"
)

type GcpNfsVolumeBackupSource struct {
	// GcpNfsVolumeRef specifies the GcpNfsVolume resource that a backup has to be made of.
	// +kubebuilder:validation:Required
	Volume GcpNfsVolumeRef `json:"volume"`
}

type GcpNfsVolumeRef struct {
	// Name specifies the name of the GcpNfsVolume resource that a backup has to be made of.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace specified the namespace of the GcpNfsVolume resource that a backup has to be made of.
	// If not specified then namespace of the GcpNfsVolumeBackup is used.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

func (v *GcpNfsVolumeRef) ToNamespacedName(fallbackNamespace string) types.NamespacedName {
	ns := v.Namespace
	if len(ns) == 0 {
		ns = fallbackNamespace
	}
	return types.NamespacedName{
		Namespace: ns,
		Name:      v.Name,
	}
}

// GcpNfsVolumeBackupSpec defines the desired state of GcpNfsVolumeBackup
type GcpNfsVolumeBackupSpec struct {

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="source is immutable."
	Source GcpNfsVolumeBackupSource `json:"source"`

	// GCP Region Name (as specified in https://cloud.google.com/filestore/docs/regions) where this backup should be created.
	// If not specified, region of the GcpNfsVolume is used for the backup.
	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Location is immutable."
	Location string `json:"location"`

	// AccessibleFrom is an array of shootNames or subaccountIds that would have access to the backup for restore.
	// "ALL" is also accepted as a value to allow access from all shoots in the same global account and gcp project. "ALL" cannot be used in combination with other values.
	// +optional
	// +listType=set
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	// +kubebuilder:validation:XValidation:rule="(self.all(x, x == 'ALL') || self.all(x, x != 'ALL'))", message="The value 'ALL' cannot be combined with other values."
	AccessibleFrom []string `json:"accessibleFrom,omitempty"`
}

// GcpNfsVolumeBackupStatus defines the observed state of GcpNfsVolumeBackup
type GcpNfsVolumeBackupStatus struct {
	State GcpNfsBackupState `json:"state,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Operation Identifier to track the Hyperscaler Restore Operation
	// +optional
	OpIdentifier string `json:"opIdentifier,omitempty"`

	// +optional
	Id string `json:"id,omitempty"`

	// +optional
	Location string `json:"location,omitempty"`

	// Capacity
	// +optional
	Capacity resource.Quantity `json:"capacity"`

	// LastCapacityUpdate specifies the time when the last time backup size got updated
	// +optional
	LastCapacityUpdate *metav1.Time `json:"lastCapacityUpdate,omitempty"`

	// Comma separated list that reflects the AccessibleFrom field in spec upon the last successful reconciliation
	AccessibleFrom string `json:"accessibleFrom,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="GCP NFS Volume",type="string",JSONPath=".spec.source.volume.name"
// +kubebuilder:printcolumn:name="Location",type="string",JSONPath=".status.location"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// GcpNfsVolumeBackup is the Schema for the gcpnfsvolumebackups API
type GcpNfsVolumeBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpNfsVolumeBackupSpec   `json:"spec,omitempty"`
	Status GcpNfsVolumeBackupStatus `json:"status,omitempty"`
}

func (in *GcpNfsVolumeBackup) State() string {
	return string(in.Status.State)
}

func (in *GcpNfsVolumeBackup) SetState(v string) {
	in.Status.State = GcpNfsBackupState(v)
}

func (in *GcpNfsVolumeBackup) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *GcpNfsVolumeBackup) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *GcpNfsVolumeBackup) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (in *GcpNfsVolumeBackup) SpecificToProviders() []string {
	return []string{"gcp"}
}

//+kubebuilder:object:root=true

// GcpNfsVolumeBackupList contains a list of GcpNfsVolumeBackup
type GcpNfsVolumeBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpNfsVolumeBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpNfsVolumeBackup{}, &GcpNfsVolumeBackupList{})
}

func (in *GcpNfsVolumeBackup) CloneForPatchStatus() client.Object {
	return &GcpNfsVolumeBackup{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GcpNfsVolumeBackup",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}
