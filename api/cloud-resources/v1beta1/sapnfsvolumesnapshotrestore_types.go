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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SapNfsVolumeSnapshotRestoreSpec defines the desired state of SapNfsVolumeSnapshotRestore
type SapNfsVolumeSnapshotRestoreSpec struct {

	// SourceSnapshot references the SapNfsVolumeSnapshot to restore from.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="SourceSnapshot is immutable."
	SourceSnapshot corev1.ObjectReference `json:"sourceSnapshot"`

	// Destination specifies where to restore the snapshot data.
	// Exactly one of ExistingVolume or NewVolume must be set.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Destination is immutable."
	Destination SapNfsVolumeSnapshotRestoreDestination `json:"destination"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type SapNfsVolumeSnapshotRestoreDestination struct {
	// ExistingVolume references an existing SapNfsVolume to revert in-place.
	// The snapshot must be the most recent snapshot of this volume.
	// +optional
	ExistingVolume *corev1.ObjectReference `json:"existingVolume,omitempty"`

	// NewVolume defines a new SapNfsVolume to create from the snapshot.
	// +optional
	NewVolume *SapNfsVolumeSnapshotNewVolume `json:"newVolume,omitempty"`
}

type SapNfsVolumeSnapshotNewVolume struct {
	// Metadata for the new SapNfsVolume (name, labels, annotations).
	// name is required; namespace defaults to the restore's namespace.
	// +kubebuilder:validation:Required
	Metadata metav1.ObjectMeta `json:"metadata"`

	// Spec is the template for the new SapNfsVolume (same type as SapNfsVolumeSpec).
	// capacityGb must be >= the snapshot's source share size.
	// +kubebuilder:validation:Required
	Spec SapNfsVolumeSpec `json:"spec"`
}

// SapNfsVolumeSnapshotRestoreStatus defines the observed state of SapNfsVolumeSnapshotRestore
type SapNfsVolumeSnapshotRestoreStatus struct {
	// State of the restore operation.
	// +optional
	State string `json:"state,omitempty"`

	// CreatedVolume references the SapNfsVolume created (new-volume restore only).
	// +optional
	CreatedVolume *corev1.ObjectReference `json:"createdVolume,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="Source Snapshot",type="string",JSONPath=".spec.sourceSnapshot.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// SapNfsVolumeSnapshotRestore is the Schema for the sapnfsvolumesnapshotrestores API
type SapNfsVolumeSnapshotRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SapNfsVolumeSnapshotRestoreSpec   `json:"spec,omitempty"`
	Status SapNfsVolumeSnapshotRestoreStatus `json:"status,omitempty"`
}

func (in *SapNfsVolumeSnapshotRestore) State() string {
	return in.Status.State
}

func (in *SapNfsVolumeSnapshotRestore) SetState(v string) {
	in.Status.State = v
}

func (in *SapNfsVolumeSnapshotRestore) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *SapNfsVolumeSnapshotRestore) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *SapNfsVolumeSnapshotRestore) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (in *SapNfsVolumeSnapshotRestore) SpecificToProviders() []string {
	return []string{"openstack"}
}

func (in *SapNfsVolumeSnapshotRestore) CloneForPatchStatus() client.Object {
	result := &SapNfsVolumeSnapshotRestore{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SapNfsVolumeSnapshotRestore",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
	if result.Status.Conditions == nil {
		result.Status.Conditions = []metav1.Condition{}
	}
	return result
}

// +kubebuilder:object:root=true

// SapNfsVolumeSnapshotRestoreList contains a list of SapNfsVolumeSnapshotRestore
type SapNfsVolumeSnapshotRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SapNfsVolumeSnapshotRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SapNfsVolumeSnapshotRestore{}, &SapNfsVolumeSnapshotRestoreList{})
}
