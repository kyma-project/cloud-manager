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
)

// AwsNfsVolumeBackupSpec defines the one time backup of the AwsNfsVolume content.
type AwsNfsVolumeBackupSpec struct {
	// Source specifies the resource which backup is being made
	// +kubebuilder:validation:Required
	Source AwsNfsVolumeBackupSource `json:"source"`

	// Lifecycle specifies the lifecycle of the created backup
	Lifecycle AwsNfsVolumeBackupLifecycle `json:"lifecycle,omitempty"`

	// +optional
	// +kubebuilder:validation:XValidation:rule=(self == oldSelf), message="Location is immutable."
	Location string `json:"location"`
}

type AwsNfsVolumeBackupSource struct {
	// Volume specifies the AwsNfsVolume resource that a backup has to be made of.
	// +kubebuilder:validation:Required
	Volume VolumeRef `json:"volume"`
}

type VolumeRef struct {
	// Name specifies the name of the AwsNfsVolume resource that a backup has to be made of.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace specified the namespace of the AwsNfsVolume resource that a backup has to be made of.
	// If not specified then namespace of the AwsNfsVolumeBackup is used.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

func (v VolumeRef) ToNamespacedName(fallbackNamespace string) types.NamespacedName {
	ns := v.Namespace
	if len(ns) == 0 {
		ns = fallbackNamespace
	}
	return types.NamespacedName{
		Namespace: ns,
		Name:      v.Name,
	}
}

// AwsNfsVolumeBackupLifecycle contains transition specification of how long in days before a
// recovery point transitions to cold storage or is deleted. Backups transitioned
// to cold storage must be stored in cold storage for a minimum of 90 days.
// Therefore, the DeleteAfterDays “retention” setting must be 90 days greater than
// the MoveToColdStorageAfterDays setting. Once resource is created these settings
// can not be changed.
type AwsNfsVolumeBackupLifecycle struct {
	// DeleteAfterDays specifies the number of days after creation that a recovery point and resource are deleted.
	// Backups transitioned to cold storage must be stored in cold storage for a minimum of 90 days.
	// So, DeleteAfterDays must be greater than 90 days plus MoveToColdStorageAfterDays
	DeleteAfterDays *int64 `json:"deleteAfterDays,omitempty"`

	// MoveToColdStorageAfterDays specifies the number of days after creation that a recovery point is moved to
	// cold storage.
	MoveToColdStorageAfterDays *int64 `json:"moveToColdStorageAfterDays,omitempty"`
}

// AwsNfsVolumeBackupStatus defines the observed state of AwsNfsVolumeBackup
type AwsNfsVolumeBackupStatus struct {
	State string `json:"state,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// Identifier of the AWS Recovery Point
	// +optional
	Id string `json:"id,omitempty"`

	// AWS Backup Job Identifier
	// +optional
	JobId string `json:"jobId,omitempty"`

	// IdempotencyToken
	// +optional
	IdempotencyToken string `json:"idempotencyToken"`

	// Capacity
	// +optional
	Capacity resource.Quantity `json:"capacity"`

	// LastCapacityUpdate specifies the time when the last time backup size got updated
	// +optional
	LastCapacityUpdate *metav1.Time `json:"lastCapacityUpdate,omitempty"`

	// Identifier of the Remote AWS Recovery Point
	// +optional
	RemoteId string `json:"remoteId,omitempty"`

	// AWS Copy Job Identifier
	// +optional
	CopyJobId string `json:"copyJobId,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}

// AwsNfsVolumeBackup is the Schema for the awsnfsvolumebackups API
type AwsNfsVolumeBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AwsNfsVolumeBackupSpec   `json:"spec,omitempty"`
	Status AwsNfsVolumeBackupStatus `json:"status,omitempty"`
}

func (in *AwsNfsVolumeBackup) State() string {
	return in.Status.State
}

func (in *AwsNfsVolumeBackup) SetState(v string) {
	in.Status.State = v
}

func (in *AwsNfsVolumeBackup) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *AwsNfsVolumeBackup) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *AwsNfsVolumeBackup) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (in *AwsNfsVolumeBackup) SpecificToProviders() []string {
	return []string{"aws"}
}

//+kubebuilder:object:root=true

// AwsNfsVolumeBackupList contains a list of AwsNfsVolumeBackup
type AwsNfsVolumeBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AwsNfsVolumeBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AwsNfsVolumeBackup{}, &AwsNfsVolumeBackupList{})
}
