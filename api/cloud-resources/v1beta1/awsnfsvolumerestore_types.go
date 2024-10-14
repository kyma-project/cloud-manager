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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AwsNfsVolumeRestoreSpec defines the desired state of AwsNfsVolumeRestore
type AwsNfsVolumeRestoreSpec struct {
	// Source specifies the backup which is getting restored. It also indirectly specifies the backup's source volume.
	// +kubebuilder:validation:Required
	Source AwsNfsVolumeRestoreSource `json:"source"`
}

type AwsNfsVolumeRestoreSource struct {
	// Volume specifies the AwsNfsVolumeBackup resource that is restored.
	// +kubebuilder:validation:Required
	Backup BackupRef `json:"backup"`
}

type BackupRef struct {
	// Name specifies the name of the AwsNfsBackup resource.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace specifies the namespace of the AwsNfsVolumeBackup resource.
	// If not specified then namespace of the AwsNfsVolumeRestore resource is used.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

func (v BackupRef) ToNamespacedName(fallbackNamespace string) types.NamespacedName {
	ns := v.Namespace
	if len(ns) == 0 {
		ns = fallbackNamespace
	}
	return types.NamespacedName{
		Namespace: ns,
		Name:      v.Name,
	}
}

// AwsNfsVolumeRestoreStatus defines the observed state of AwsNfsVolumeRestore
type AwsNfsVolumeRestoreStatus struct {
	State string `json:"state,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// AWS Restore Job Identifier
	// +optional
	JobId string `json:"jobId,omitempty"`

	// IdempotencyToken
	// +optional
	IdempotencyToken string `json:"idempotencyToken"`

	// The directory under the root of volume where the backup is restored. Only applies to in place restores.
	// +optional
	RestoredDir string `json:"restoredDir,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:printcolumn:name="Source",type="string",JSONPath=".spec.source.backup.name"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// AwsNfsVolumeRestore is the Schema for the awsnfsvolumerestores API
type AwsNfsVolumeRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsNfsVolumeRestoreSpec   `json:"spec,omitempty"`
	Status AwsNfsVolumeRestoreStatus `json:"status,omitempty"`
}

func (in *AwsNfsVolumeRestore) State() string {
	return in.Status.State
}

func (in *AwsNfsVolumeRestore) SetState(v string) {
	in.Status.State = v
}

func (in *AwsNfsVolumeRestore) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AwsNfsVolumeRestore) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AwsNfsVolumeRestore) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (in *AwsNfsVolumeRestore) SpecificToProviders() []string {
	return []string{"aws"}
}

//+kubebuilder:object:root=true

// AwsNfsVolumeRestoreList contains a list of AwsNfsVolumeRestore
type AwsNfsVolumeRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsNfsVolumeRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsNfsVolumeRestore{}, &AwsNfsVolumeRestoreList{})
}

func (in *AwsNfsVolumeRestore) CloneForPatchStatus() client.Object {
	return &AwsNfsVolumeRestore{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AwsNfsVolumeRestore",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}
