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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

const (
	ReasonInvalidCronExpression = "InvalidCronExpression"
	ReasonTimeParseError        = "TimeParseError"
	ReasonScheduleError         = "ScheduleError"
	ReasonNfsVolumeNotFound     = "NfsVolumeNotFound"
	ReasonNfsVolumeNotReady     = "NfsVolumeNotReady"
	ReasonBackupCreateFailed    = "BackupCreateFailed"
	ReasonBackupListFailed      = "BackupListFailed"
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
	// +kubebuilder:validation:Enum=Processing;Pending;Suspended;Active;Done;Error
	State string `json:"state,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// NextRunTimes contains the times when the next backup will be created
	// +optional
	NextRunTimes []string `json:"nextRunTimes,omitempty"`

	//NextDeleteTimes contains the map of backup objects and
	//their expected deletion times (calculated based on MaxRetentionDays).
	// +optional
	NextDeleteTimes map[string]string `json:"nextDeleteTimes,omitempty"`

	// LastCreateRun specifies the time when the last backup was created
	// +optional
	LastCreateRun *metav1.Time `json:"lastCreateRun,omitempty"`

	// LastCreatedBackup contains the object reference of the backup object created during last run.
	// +optional
	LastCreatedBackup corev1.ObjectReference `json:"lastCreatedBackup,omitempty"`

	// LastDeleteRun specifies the time when the backups exceeding the retention period were deleted
	// +optional
	LastDeleteRun *metav1.Time `json:"lastDeleteRun,omitempty"`

	// LastDeletedBackups contains the object references of the backup object deleted during last run.
	// +optional
	LastDeletedBackups []corev1.ObjectReference `json:"lastDeletedBackups,omitempty"`

	// Schedule specifies the cron expression of the current active schedule
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// BackupIndex specifies the current index of the backup created by this schedule
	// +kubebuilder:default=0
	BackupIndex int `json:"backupIndex,omitempty"`
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

func (sc *NfsBackupSchedule) Conditions() *[]metav1.Condition {
	return &sc.Status.Conditions
}

func (sc *NfsBackupSchedule) GetObjectMeta() *metav1.ObjectMeta {
	return &sc.ObjectMeta
}

func (sc *NfsBackupSchedule) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (sc *NfsBackupSchedule) SpecificToProviders() []string {
	return []string{"gcp"}
}

func (sc *NfsBackupSchedule) State() string {
	return sc.Status.State
}
func (sc *NfsBackupSchedule) SetState(state string) {
	sc.Status.State = state
}
func (sc *NfsBackupSchedule) GetSourceRef() corev1.ObjectReference {
	return sc.Spec.NfsVolumeRef
}
func (sc *NfsBackupSchedule) SetSourceRef(ref corev1.ObjectReference) {
	sc.Spec.NfsVolumeRef = ref
}
func (sc *NfsBackupSchedule) GetSchedule() string {
	return sc.Spec.Schedule
}
func (sc *NfsBackupSchedule) SetSchedule(schedule string) {
	sc.Spec.Schedule = schedule
}
func (sc *NfsBackupSchedule) GetPrefix() string {
	return sc.Spec.Prefix
}
func (sc *NfsBackupSchedule) SetPrefix(prefix string) {
	sc.Spec.Prefix = prefix
}
func (sc *NfsBackupSchedule) GetStartTime() *metav1.Time {
	return sc.Spec.StartTime
}
func (sc *NfsBackupSchedule) SetStartTime(start *metav1.Time) {
	sc.Spec.StartTime = start
}
func (sc *NfsBackupSchedule) GetEndTime() *metav1.Time {
	return sc.Spec.EndTime
}
func (sc *NfsBackupSchedule) SetEndTime(end *metav1.Time) {
	sc.Spec.EndTime = end
}
func (sc *NfsBackupSchedule) GetMaxRetentionDays() int {
	return sc.Spec.MaxRetentionDays
}
func (sc *NfsBackupSchedule) SetMaxRetentionDays(days int) {
	sc.Spec.MaxRetentionDays = days
}
func (sc *NfsBackupSchedule) GetSuspend() bool {
	return sc.Spec.Suspend
}
func (sc *NfsBackupSchedule) SetSuspend(suspend bool) {
	sc.Spec.Suspend = suspend
}

func (sc *NfsBackupSchedule) GetNextRunTimes() []string {
	return sc.Status.NextRunTimes
}
func (sc *NfsBackupSchedule) SetNextRunTimes(times []string) {
	sc.Status.NextRunTimes = times
}
func (sc *NfsBackupSchedule) GetNextDeleteTimes() map[string]string {
	return sc.Status.NextDeleteTimes
}
func (sc *NfsBackupSchedule) SetNextDeleteTimes(times map[string]string) {
	sc.Status.NextDeleteTimes = times
}
func (sc *NfsBackupSchedule) GetLastCreateRun() *metav1.Time {
	return sc.Status.LastCreateRun
}
func (sc *NfsBackupSchedule) SetLastCreateRun(time *metav1.Time) {
	sc.Status.LastCreateRun = time
}
func (sc *NfsBackupSchedule) GetLastCreatedBackup() corev1.ObjectReference {
	return sc.Status.LastCreatedBackup
}
func (sc *NfsBackupSchedule) SetLastCreatedBackup(obj corev1.ObjectReference) {
	sc.Status.LastCreatedBackup = obj
}
func (sc *NfsBackupSchedule) GetLastDeleteRun() *metav1.Time {
	return sc.Status.LastDeleteRun
}
func (sc *NfsBackupSchedule) SetLastDeleteRun(time *metav1.Time) {
	sc.Status.LastDeleteRun = time
}
func (sc *NfsBackupSchedule) GetLastDeletedBackups() []corev1.ObjectReference {
	return sc.Status.LastDeletedBackups
}
func (sc *NfsBackupSchedule) SetLastDeletedBackups(objs []corev1.ObjectReference) {
	sc.Status.LastDeletedBackups = objs
}
func (sc *NfsBackupSchedule) GetActiveSchedule() string {
	return sc.Status.Schedule
}
func (sc *NfsBackupSchedule) SetActiveSchedule(schedule string) {
	sc.Status.Schedule = schedule
}
func (sc *NfsBackupSchedule) GetBackupIndex() int {
	return sc.Status.BackupIndex
}
func (sc *NfsBackupSchedule) SetBackupIndex(index int) {
	sc.Status.BackupIndex = index
}
func (sc *NfsBackupSchedule) GetList() client.ObjectList {
	return &NfsBackupScheduleList{}
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
