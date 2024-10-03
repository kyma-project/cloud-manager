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
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GcpNfsVolumeBackupRef struct {
	// Name specifies the name of the GcpNfsVolumeBackup resource that would be restored.
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Namespace specifies the namespace of the GcpNfsVolumeBackup resource that would be restored.
	// +kubebuilder:validation:Required
	Namespace string `json:"namespace"`
}

func (v *GcpNfsVolumeBackupRef) ToNamespacedName(fallbackNamespace string) types.NamespacedName {
	ns := v.Namespace
	if len(ns) == 0 {
		ns = fallbackNamespace
	}
	return types.NamespacedName{
		Namespace: ns,
		Name:      v.Name,
	}
}

// GcpNfsVolumeRestoreSpec defines the desired spec of GcpNfsVolumeRestore
type GcpNfsVolumeRestoreSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Source is immutable."
	Source GcpNfsVolumeRestoreSource `json:"source"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Destination is immutable."
	Destination GcpNfsVolumeRestoreDestination `json:"destination"`
}

type GcpNfsVolumeRestoreSource struct {
	// GcpNfsVolumeBackupRef specifies the GcpNfsVolumeBackup resource that would be restored
	// +kubebuilder:validation:Required
	Backup GcpNfsVolumeBackupRef `json:"backup"`
}

type GcpNfsVolumeRestoreDestination struct {
	// GcpNfsVolumeRef specifies the GcpNfsVolume resource that a backup is restored on.
	// +kubebuilder:validation:Required
	Volume GcpNfsVolumeRef `json:"volume"`
}

// GcpNfsVolumeRestoreStatus defines the observed state of GcpNfsVolumeRestore
type GcpNfsVolumeRestoreStatus struct {
	// +kubebuilder:validation:Enum=Processing;InProgress;Done;Failed;Error
	State string `json:"state,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Operation Identifier to track the Hyperscaler Restore Operation
	// +optional
	OpIdentifier string `json:"opIdentifier,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Source",type="string",JSONPath=".spec.source.backup.name"
// +kubebuilder:printcolumn:name="Destination",type="string",JSONPath=".spec.destination.volume.name"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// GcpNfsVolumeRestore is the Schema for the gcpnfsvolumerestores API
type GcpNfsVolumeRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpNfsVolumeRestoreSpec   `json:"spec,omitempty"`
	Status GcpNfsVolumeRestoreStatus `json:"status,omitempty"`
}

func (in *GcpNfsVolumeRestore) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *GcpNfsVolumeRestore) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *GcpNfsVolumeRestore) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (in *GcpNfsVolumeRestore) SpecificToProviders() []string {
	return []string{"gcp"}
}

//+kubebuilder:object:root=true

// GcpNfsVolumeRestoreList contains a list of GcpNfsVolumeRestore
type GcpNfsVolumeRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpNfsVolumeRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpNfsVolumeRestore{}, &GcpNfsVolumeRestoreList{})
}

func (in *GcpNfsVolumeRestore) CloneForPatchStatus() client.Object {
	return &GcpNfsVolumeRestore{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GcpNfsVolumeRestore",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}
