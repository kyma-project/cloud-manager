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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ReasonInvalidCronExpression = "InvalidCronExpression"
	ReasonTimeParseError        = "TimeParseError"
	ReasonScheduleError         = "ScheduleError"
	ReasonNfsVolumeNotFound     = "NfsVolumeNotFound"
	ReasonNfsVolumeNotReady     = "NfsVolumeNotReady"
	ReasonBackupCreateFailed    = "BackupCreateFailed"
)

// NfsBackupScheduleSpec defines the desired state of NfsBackupSchedule
type NfsBackupScheduleSpec struct {

	// NfsVolumeRef specifies the NfsVolume resource that a backup has to be made of.
	// +kubebuilder:validation:Required
	NfsVolumeRef corev1.ObjectReference `json:"nfsVolumeRef"`

	// Location specifies the location where the backup has to be stored.
	// +optional
	Location string `json:"location"`

	// Cron expression of the schedule, e.g. "0 0 * * *" for daily at midnight
	// If not provided, backup will be taken once on the specified start time.
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// Prefix for the backup name.
	// If not provided, schedule name will be used as prefix
	// +optional
	Prefix string `json:"prefix,omitempty"`

	// StartTime specifies the time when the backup should start
	// If not provided, schedule will start immediately
	// +optional
	// +kubebuilder:validation:Format=date-time
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// EndTime specifies the time when the backup should end
	// If not provided, schedule will run indefinitely
	// +optional
	// +kubebuilder:validation:Format=date-time
	EndTime *metav1.Time `json:"endTime,omitempty"`

	// MaxRetentionDays specifies the maximum number of days to retain the backup
	// If not provided, backup will be retained indefinitely
	// +optional
	MaxRetentionDays int `json:"maxRetentionDays,omitempty"`

	// Suspend specifies whether the schedule should be suspended
	// By default, suspend will be false
	// +kubebuilder:default=false
	Suspend bool `json:"suspend,omitempty"`
}

// NfsBackupScheduleStatus defines the observed state of NfsBackupSchedule
type NfsBackupScheduleStatus struct {
	// +kubebuilder:validation:Enum=Processing;Pending;Suspended;Active;Error
	State string `json:"state,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// NextRunTimes contains the time when the next backup will be created
	// +optional
	NextRunTimes []string `json:"nextRunTimes,omitempty"`

	// LastCreateRun specifies the time when the last backup was created
	// +optional
	LastCreateRun *metav1.Time `json:"lastCreateRun,omitempty"`

	// LastDeleteRun specifies the time when the backups exceeding the retention period were deleted
	// +optional
	LastDeleteRun *metav1.Time `json:"lastDeleteRun,omitempty"`

	// Schedule specifies the cron expression of the current active schedule
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// BackupIndex specifies the current index of the backup created by this schedule
	// +kubebuilder:default=0
	BackupIndex int `json:"backupIndex,omitempty"`

	// Backups specifies the list of backups created by this schedule
	// +optional
	Backups []corev1.ObjectReference `json:"backups,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule"
// +kubebuilder:printcolumn:name="Last Run Time",type="date",JSONPath=".status.lastCreateRun"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// NfsBackupSchedule is the Schema for the nfsbackupschedules API
type NfsBackupSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NfsBackupScheduleSpec   `json:"spec,omitempty"`
	Status NfsBackupScheduleStatus `json:"status,omitempty"`
}

func (in *NfsBackupSchedule) Conditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

func (in *NfsBackupSchedule) GetObjectMeta() *metav1.ObjectMeta {
	return &in.ObjectMeta
}

func (in *NfsBackupSchedule) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (in *NfsBackupSchedule) SpecificToProviders() []string {
	return nil
}

//+kubebuilder:object:root=true

// NfsBackupScheduleList contains a list of NfsBackupSchedule
type NfsBackupScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NfsBackupSchedule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NfsBackupSchedule{}, &NfsBackupScheduleList{})
}
