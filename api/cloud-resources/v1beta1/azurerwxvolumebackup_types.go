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
)

type AzureRwxBackupState string

const (
	AzureRwxBackupReady AzureRwxBackupState = "Ready"

	AzureRwxBackupError AzureRwxBackupState = "Error"

	AzureRwxBackupCreating AzureRwxBackupState = "Creating"

	AzureRwxBackupDeleting AzureRwxBackupState = "Deleting"

	AzureRwxBackupDeleted AzureRwxBackupState = "Deleted"

	AzureRwxBackupFailed AzureRwxBackupState = "Failed"
)

type PvcRef struct {
	// Name speicfies the name of the PVC that a backup has to be made of.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace specified the namespace of the AzureRwxVolume resource that a backup has to be made of.
	// If not specified then namespace of the AzureRwxVolumeBackup is used.
	// +optional
	Namespace string `json:"namespace"`
}

func (v *PvcRef) ToNamespacedName(fallbackNamespace string) types.NamespacedName {
	ns := v.Namespace
	if len(ns) == 0 {
		ns = fallbackNamespace
	}
	return types.NamespacedName{
		Namespace: ns,
		Name:      v.Name,
	}
}

type PvcSource struct {
	Pvc PvcRef `json:"pvc"`
}

// AzureRwxVolumeBackupSpec defines the desired state of AzureRwxVolumeBackup
type AzureRwxVolumeBackupSpec struct {

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="source is immutable."
	Source PvcSource `json:"source"`

	Location string `json:"location"`
}

// AzureRwxVolumeBackupStatus defines the observed state of AzureRwxVolumeBackup
type AzureRwxVolumeBackupStatus struct {
	State AzureRwxBackupState `json:"state,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Operation Identifier to track the Hyperscaler Creation Operation
	// +optional
	OpIdentifier string `json:"opIdentifier,omitempty"`

	// +optional
	Id string `json:"id,omitempty"`

	// RecoveryPointId specifies the corresponding snapshot used for restore
	// +optional
	RecoveryPointId string `json:"recoveryPointId,omitempty"`

	// StorageAccountPath specifies the Azure Storage Account path
	// +optional
	StorageAccountPath string `json:"storageAccountPath,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AzureRwxVolumeBackup is the Schema for the azurerwxvolumebackups API
type AzureRwxVolumeBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureRwxVolumeBackupSpec   `json:"spec,omitempty"`
	Status AzureRwxVolumeBackupStatus `json:"status,omitempty"`
}

func (bu *AzureRwxVolumeBackup) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (bu *AzureRwxVolumeBackup) SpecificToProviders() []string {
	return []string{"azure"}
}

// +kubebuilder:object:root=true

// AzureRwxVolumeBackupList contains a list of AzureRwxVolumeBackup
type AzureRwxVolumeBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureRwxVolumeBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureRwxVolumeBackup{}, &AzureRwxVolumeBackupList{})
}
