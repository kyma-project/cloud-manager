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

// GcpNfsBackupScheduleSpec defines the desired state of GcpNfsBackupSchedule
type GcpNfsBackupScheduleSpec struct {

	// NfsVolumeRef specifies the SourceRef resource that a backup has to be made of.
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
	// If not provided, it will be defaulted to 375 days.
	// If the DeleteCascade is true for this schedule,
	// then all the backups will be deleted when the schedule is deleted irrespective of the MaxRetentionDay configuration.
	// +optional
	// +kubebuilder:default=375
	// +kubebuilder:validation:Minimum=1
	MaxRetentionDays int `json:"maxRetentionDays,omitempty"`

	// MaxReadyBackups specifies the maximum number of backups in "Ready" state to be retained.
	// If not provided, it will be defaulted to 100 active backups.
	// If the DeleteCascade is true for this schedule,
	// then all the backups will be deleted when the schedule is deleted irrespective of the MaxRetentionDay configuration.
	// +optional
	// +kubebuilder:default=100
	// +kubebuilder:validation:Minimum=1
	MaxReadyBackups int `json:"maxReadyBackups,omitempty"`

	// MaxFailedBackups specifies the maximum number of backups in "Failed" state to be retained.
	// If not provided, it will be defaulted to 5 failed backups.
	// If the DeleteCascade is true for this schedule,
	// then all the backups will be deleted when the schedule is deleted irrespective of the MaxRetentionDay configuration.
	// +optional
	// +kubebuilder:default=5
	// +kubebuilder:validation:Minimum=1
	MaxFailedBackups int `json:"maxFailedBackups,omitempty"`

	// Suspend specifies whether the schedule should be suspended
	// By default, suspend will be false
	// +kubebuilder:default=false
	Suspend bool `json:"suspend,omitempty"`

	// DeleteCascade specifies whether to cascade delete the backups when this schedule is deleted.
	// By default, deleteCascade will be false
	// +kubebuilder:default=false
	DeleteCascade bool `json:"deleteCascade,omitempty"`
}

// GcpNfsBackupScheduleStatus defines the observed state of GcpNfsBackupSchedule
type GcpNfsBackupScheduleStatus struct {
	// +kubebuilder:validation:Enum=Processing;Pending;Suspended;Active;Done;Error;Deleting
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

	//BackupCount specifies the number of backups currently present in the system
	// +kubebuilder:default=0
	BackupCount int `json:"backupCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule"
// +kubebuilder:printcolumn:name="Last Run Time",type="date",JSONPath=".status.lastCreateRun"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// GcpNfsBackupSchedule is the Schema for the GcpNfsBackupSchedules API
type GcpNfsBackupSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GcpNfsBackupScheduleSpec   `json:"spec,omitempty"`
	Status GcpNfsBackupScheduleStatus `json:"status,omitempty"`
}

func (sc *GcpNfsBackupSchedule) Conditions() *[]metav1.Condition {
	return &sc.Status.Conditions
}

func (sc *GcpNfsBackupSchedule) GetObjectMeta() *metav1.ObjectMeta {
	return &sc.ObjectMeta
}

func (sc *GcpNfsBackupSchedule) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (sc *GcpNfsBackupSchedule) SpecificToProviders() []string {
	return []string{"gcp"}
}

func (sc *GcpNfsBackupSchedule) State() string {
	return sc.Status.State
}
func (sc *GcpNfsBackupSchedule) SetState(state string) {
	sc.Status.State = state
}
func (sc *GcpNfsBackupSchedule) GetSourceRef() corev1.ObjectReference {
	return sc.Spec.NfsVolumeRef
}
func (sc *GcpNfsBackupSchedule) SetSourceRef(ref corev1.ObjectReference) {
	sc.Spec.NfsVolumeRef = ref
}
func (sc *GcpNfsBackupSchedule) GetSchedule() string {
	return sc.Spec.Schedule
}
func (sc *GcpNfsBackupSchedule) SetSchedule(schedule string) {
	sc.Spec.Schedule = schedule
}
func (sc *GcpNfsBackupSchedule) GetPrefix() string {
	return sc.Spec.Prefix
}
func (sc *GcpNfsBackupSchedule) SetPrefix(prefix string) {
	sc.Spec.Prefix = prefix
}
func (sc *GcpNfsBackupSchedule) GetStartTime() *metav1.Time {
	return sc.Spec.StartTime
}
func (sc *GcpNfsBackupSchedule) SetStartTime(start *metav1.Time) {
	sc.Spec.StartTime = start
}
func (sc *GcpNfsBackupSchedule) GetEndTime() *metav1.Time {
	return sc.Spec.EndTime
}
func (sc *GcpNfsBackupSchedule) SetEndTime(end *metav1.Time) {
	sc.Spec.EndTime = end
}
func (sc *GcpNfsBackupSchedule) GetMaxRetentionDays() int {
	return sc.Spec.MaxRetentionDays
}
func (sc *GcpNfsBackupSchedule) SetMaxRetentionDays(days int) {
	sc.Spec.MaxRetentionDays = days
}
func (sc *GcpNfsBackupSchedule) GetSuspend() bool {
	return sc.Spec.Suspend
}
func (sc *GcpNfsBackupSchedule) SetSuspend(suspend bool) {
	sc.Spec.Suspend = suspend
}

func (sc *GcpNfsBackupSchedule) GetDeleteCascade() bool {
	return sc.Spec.DeleteCascade
}

func (sc *GcpNfsBackupSchedule) SetDeleteCascade(cascade bool) {
	sc.Spec.DeleteCascade = cascade
}

func (sc *GcpNfsBackupSchedule) GetMaxReadyBackups() int {
	return sc.Spec.MaxReadyBackups
}
func (sc *GcpNfsBackupSchedule) SetMaxReadyBackups(count int) {
	sc.Spec.MaxReadyBackups = count
}

func (sc *GcpNfsBackupSchedule) GetMaxFailedBackups() int {
	return sc.Spec.MaxFailedBackups
}
func (sc *GcpNfsBackupSchedule) SetMaxFailedBackups(count int) {
	sc.Spec.MaxFailedBackups = count
}

func (sc *GcpNfsBackupSchedule) GetNextRunTimes() []string {
	return sc.Status.NextRunTimes
}
func (sc *GcpNfsBackupSchedule) SetNextRunTimes(times []string) {
	sc.Status.NextRunTimes = times
}
func (sc *GcpNfsBackupSchedule) GetNextDeleteTimes() map[string]string {
	return sc.Status.NextDeleteTimes
}
func (sc *GcpNfsBackupSchedule) SetNextDeleteTimes(times map[string]string) {
	sc.Status.NextDeleteTimes = times
}
func (sc *GcpNfsBackupSchedule) GetLastCreateRun() *metav1.Time {
	return sc.Status.LastCreateRun
}
func (sc *GcpNfsBackupSchedule) SetLastCreateRun(time *metav1.Time) {
	sc.Status.LastCreateRun = time
}
func (sc *GcpNfsBackupSchedule) GetLastCreatedBackup() corev1.ObjectReference {
	return sc.Status.LastCreatedBackup
}
func (sc *GcpNfsBackupSchedule) SetLastCreatedBackup(obj corev1.ObjectReference) {
	sc.Status.LastCreatedBackup = obj
}
func (sc *GcpNfsBackupSchedule) GetLastDeleteRun() *metav1.Time {
	return sc.Status.LastDeleteRun
}
func (sc *GcpNfsBackupSchedule) SetLastDeleteRun(time *metav1.Time) {
	sc.Status.LastDeleteRun = time
}
func (sc *GcpNfsBackupSchedule) GetLastDeletedBackups() []corev1.ObjectReference {
	return sc.Status.LastDeletedBackups
}
func (sc *GcpNfsBackupSchedule) SetLastDeletedBackups(objs []corev1.ObjectReference) {
	sc.Status.LastDeletedBackups = objs
}
func (sc *GcpNfsBackupSchedule) GetActiveSchedule() string {
	return sc.Status.Schedule
}
func (sc *GcpNfsBackupSchedule) SetActiveSchedule(schedule string) {
	sc.Status.Schedule = schedule
}
func (sc *GcpNfsBackupSchedule) GetBackupIndex() int {
	return sc.Status.BackupIndex
}
func (sc *GcpNfsBackupSchedule) SetBackupIndex(index int) {
	sc.Status.BackupIndex = index
}
func (sc *GcpNfsBackupSchedule) GetBackupCount() int {
	return sc.Status.BackupCount
}
func (sc *GcpNfsBackupSchedule) SetBackupCount(count int) {
	sc.Status.BackupCount = count
}

//+kubebuilder:object:root=true

// GcpNfsBackupScheduleList contains a list of GcpNfsBackupSchedule
type GcpNfsBackupScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GcpNfsBackupSchedule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GcpNfsBackupSchedule{}, &GcpNfsBackupScheduleList{})
}

func (sc *GcpNfsBackupSchedule) CloneForPatchStatus() client.Object {
	return &GcpNfsBackupSchedule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "GcpNfsBackupSchedule",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sc.Namespace,
			Name:      sc.Name,
		},
		Status: sc.Status,
	}
}
