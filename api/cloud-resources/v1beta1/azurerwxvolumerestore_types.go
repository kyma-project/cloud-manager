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

type AzureRwxVolumeBackupRef struct {
	// Name specifies the name of the AzureRwxVolumeBackup resource that would be restored.
	// +kubebuilder:validation:Required
	Name string `json:"name"`
	// Namespace specifies the namespace of the AzureRwxVolumeBackup resource that would be restored.
	// +optional
	Namespace string `json:"namespace"`
}

func (v *AzureRwxVolumeBackupRef) ToNamespacedName(fallbackNamespace string) types.NamespacedName {
	ns := v.Namespace
	if len(ns) == 0 {
		ns = fallbackNamespace
	}
	return types.NamespacedName{
		Namespace: ns,
		Name:      v.Name,
	}
}

type AzureRwxVolumeRestoreSource struct {
	// AzureRwxVolumeBackupRef specifies the AzureRwxVolumeBackup resource that would be restored
	// +kubebuilder:validation:Required
	Backup AzureRwxVolumeBackupRef `json:"backup"`
}

// AzureRwxVolumeRestoreSpec defines the desired state of AzureRwxVolumeRestore
type AzureRwxVolumeRestoreSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Source is immutable."
	Source AzureRwxVolumeRestoreSource `json:"source"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Destination is immutable."
	Destination PvcSource `json:"destination"`
}

// AzureRwxVolumeRestoreStatus defines the observed state of AzureRwxVolumeRestore
type AzureRwxVolumeRestoreStatus struct {
	// +kubebuilder:validation:Enum=Processing;InProgress;Done;Failed;Error
	State string `json:"state,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// The directory under the root of volume where the backup is restored.
	// +optional
	RestoredDir string `json:"restoredDir,omitempty"`

	// The time when the restore operation is about to start.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// +optional
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

// AzureRwxVolumeRestore is the Schema for the azurerwxvolumerestores API
type AzureRwxVolumeRestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureRwxVolumeRestoreSpec   `json:"spec,omitempty"`
	Status AzureRwxVolumeRestoreStatus `json:"status,omitempty"`
}

func (in *AzureRwxVolumeRestore) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AzureRwxVolumeRestore) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AzureRwxVolumeRestore) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (in *AzureRwxVolumeRestore) SpecificToProviders() []string {
	return []string{"azure"}
}

//+kubebuilder:object:root=true

// AzureRwxVolumeRestoreList contains a list of AzureRwxVolumeRestore
type AzureRwxVolumeRestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureRwxVolumeRestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureRwxVolumeRestore{}, &AzureRwxVolumeRestoreList{})
}

func (in *AzureRwxVolumeRestore) CloneForPatchStatus() client.Object {
	return &AzureRwxVolumeRestore{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AzureRwxVolumeRestore",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: in.Namespace,
			Name:      in.Name,
		},
		Status: in.Status,
	}
}

// additional condition reasons
const (
	ConditionReasonMissingRwxVolumeBackup          = "MissingRwxVolumeBackup"
	ConditionReasonRwxVolumeBackupNotReady         = "RwxVolumeBackupNotReady"
	ConditionReasonPvcNotFound                     = "PvcNotFound"
	ConditionReasonPvcNotBound                     = "PvcNotBound"
	ConditionReasonPvNotFound                      = "PvNotFound"
	ConditionReasonPvNotBound                      = "PvNotBound"
	ConditionReasonInvalidProvisioner              = "InvalidProvisioner"
	ConditionReasonInvalidVolumeHandle             = "InvalidVolumeHandle"
	ConditionReasonInvalidRecoveryPointId          = "InvalidRecoveryPointId"
	ConditionReasonInvalidStorageAccountPath       = "InvalidStorageAccountPath"
	ConditionReasonRestoreJobFailed                = "RestoreJobFailed"
	ConditionReasonRestoreJobCancelled             = "RestoreJobCancelled"
	ConditionReasonRestoreJobInvalidStatus         = "RestoreJobInvalidStatus"
	ConditionReasonRestoreJobCompletedWithWarnings = "RestoreJobCompletedWithWarnings"
	ConditionReasonRestoreJobNotFound              = "RestoreJobNotFound"
)
