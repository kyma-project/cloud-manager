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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SapNfsVolumeSnapshotSpec defines the desired state of SapNfsVolumeSnapshot
type SapNfsVolumeSnapshotSpec struct {

	// SourceVolume references the SapNfsVolume to snapshot.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="SourceVolume is immutable."
	SourceVolume SapNfsVolumeRef `json:"sourceVolume"`

	// DeleteAfterDays specifies the number of days after which the snapshot
	// will be automatically deleted. 0 means no automatic deletion.
	// +optional
	DeleteAfterDays int `json:"deleteAfterDays,omitempty"`
}

// SapNfsVolumeRef references a SapNfsVolume resource.
type SapNfsVolumeRef struct {
	// Name specifies the name of the SapNfsVolume resource.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace specifies the namespace of the SapNfsVolume resource.
	// If not specified then namespace of the parent resource is used.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// SapNfsVolumeSnapshotStatus defines the observed state of SapNfsVolumeSnapshot
type SapNfsVolumeSnapshotStatus struct {
	// State of the snapshot lifecycle.
	// +optional
	State string `json:"state,omitempty"`

	// Id is the internal snapshot identifier.
	// +optional
	Id string `json:"id,omitempty"`

	// OpenstackId is the Manila snapshot UUID.
	// +optional
	OpenstackId string `json:"openstackId,omitempty"`

	// SizeGb is the snapshot size in GiB as reported by Manila.
	// +optional
	SizeGb int `json:"sizeGb,omitempty"`

	// ShareId is the Manila share UUID the snapshot belongs to.
	// +optional
	ShareId string `json:"shareId,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="Source Volume",type="string",JSONPath=".spec.sourceVolume.name"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// SapNfsVolumeSnapshot is the Schema for the sapnfsvolumesnapshots API
type SapNfsVolumeSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SapNfsVolumeSnapshotSpec   `json:"spec,omitempty"`
	Status SapNfsVolumeSnapshotStatus `json:"status,omitempty"`
}

func (in *SapNfsVolumeSnapshot) State() string {
	return in.Status.State
}

func (in *SapNfsVolumeSnapshot) SetState(v string) {
	in.Status.State = v
}

func (in *SapNfsVolumeSnapshot) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *SapNfsVolumeSnapshot) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *SapNfsVolumeSnapshot) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (in *SapNfsVolumeSnapshot) SpecificToProviders() []string {
	return []string{"openstack"}
}

func (in *SapNfsVolumeSnapshot) CloneForPatchStatus() client.Object {
	result := &SapNfsVolumeSnapshot{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SapNfsVolumeSnapshot",
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

// SapNfsVolumeSnapshotList contains a list of SapNfsVolumeSnapshot
type SapNfsVolumeSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SapNfsVolumeSnapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SapNfsVolumeSnapshot{}, &SapNfsVolumeSnapshotList{})
}
