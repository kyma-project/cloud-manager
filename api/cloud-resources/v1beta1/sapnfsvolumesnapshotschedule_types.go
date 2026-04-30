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

// SapNfsVolumeSnapshotScheduleSpec defines the desired state of SapNfsVolumeSnapshotSchedule
type SapNfsVolumeSnapshotScheduleSpec struct {

	// Schedule is a cron expression. If empty, creates a one-time snapshot.
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// Prefix is used as a prefix for snapshot names.
	// If not provided, schedule name will be used as prefix.
	// +optional
	Prefix string `json:"prefix,omitempty"`

	// StartTime specifies the time when the schedule should start creating snapshots.
	// If not provided, schedule will start immediately.
	// +optional
	// +kubebuilder:validation:Format=date-time
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// EndTime specifies the time when the schedule should stop creating snapshots.
	// If not provided, schedule will run indefinitely.
	// +optional
	// +kubebuilder:validation:Format=date-time
	EndTime *metav1.Time `json:"endTime,omitempty"`

	// MaxRetentionDays specifies the maximum number of days to retain snapshots.
	// Stamped as deleteAfterDays on each created snapshot.
	// If not provided, it will be defaulted to 375 days.
	// +optional
	// +kubebuilder:default=375
	// +kubebuilder:validation:Minimum=1
	MaxRetentionDays int `json:"maxRetentionDays,omitempty"`

	// MaxReadySnapshots is the maximum number of Ready snapshots to keep.
	// +optional
	// +kubebuilder:default=50
	// +kubebuilder:validation:Minimum=1
	MaxReadySnapshots int `json:"maxReadySnapshots,omitempty"`

	// MaxFailedSnapshots is the maximum number of Failed snapshots to keep.
	// +optional
	// +kubebuilder:default=5
	// +kubebuilder:validation:Minimum=1
	MaxFailedSnapshots int `json:"maxFailedSnapshots,omitempty"`

	// Suspend stops the schedule from creating new snapshots.
	// +kubebuilder:default=false
	Suspend bool `json:"suspend,omitempty"`

	// DeleteCascade specifies whether to cascade delete the snapshots when this schedule is deleted.
	// +kubebuilder:default=false
	DeleteCascade bool `json:"deleteCascade,omitempty"`

	// Template defines the SapNfsVolumeSnapshot to create on each run.
	// +kubebuilder:validation:Required
	Template SapNfsVolumeSnapshotTemplate `json:"template"`
}

// SapNfsVolumeSnapshotTemplate defines the template for creating snapshots.
type SapNfsVolumeSnapshotTemplate struct {
	// Labels to apply to created SapNfsVolumeSnapshot objects.
	// Merged with schedule-managed labels.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations to apply to created SapNfsVolumeSnapshot objects.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Spec is the template for the SapNfsVolumeSnapshot spec.
	// +kubebuilder:validation:Required
	Spec SapNfsVolumeSnapshotSpec `json:"spec"`
}

// SapNfsVolumeSnapshotScheduleStatus defines the observed state of SapNfsVolumeSnapshotSchedule
type SapNfsVolumeSnapshotScheduleStatus struct {
	// State of the schedule.
	// +optional
	State string `json:"state,omitempty"`

	// List of status conditions
	// +optional
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// NextRunTimes contains the times when the next snapshot will be created.
	// +optional
	NextRunTimes []string `json:"nextRunTimes,omitempty"`

	// NextDeleteTimes contains the map of snapshot objects and
	// their expected deletion times (calculated based on MaxRetentionDays).
	// +optional
	NextDeleteTimes map[string]string `json:"nextDeleteTimes,omitempty"`

	// LastCreateRun specifies the time when the last snapshot was created.
	// +optional
	LastCreateRun *metav1.Time `json:"lastCreateRun,omitempty"`

	// LastCreatedBackup contains the object reference of the snapshot created during last run.
	// +optional
	LastCreatedBackup corev1.ObjectReference `json:"lastCreatedBackup,omitempty"`

	// LastDeleteRun specifies the time when the snapshots exceeding the retention period were deleted.
	// +optional
	LastDeleteRun *metav1.Time `json:"lastDeleteRun,omitempty"`

	// LastDeletedBackups contains the object references of the snapshots deleted during last run.
	// +optional
	LastDeletedBackups []corev1.ObjectReference `json:"lastDeletedBackups,omitempty"`

	// Schedule specifies the cron expression of the current active schedule.
	// +optional
	Schedule string `json:"schedule,omitempty"`

	// SnapshotIndex is the monotonically incrementing index for snapshot naming.
	// +kubebuilder:default=0
	SnapshotIndex int `json:"snapshotIndex,omitempty"`

	// BackupCount specifies the number of snapshots currently present in the system.
	// +kubebuilder:default=0
	BackupCount int `json:"backupCount,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={kyma-cloud-manager}
// +kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=".spec.schedule"
// +kubebuilder:printcolumn:name="Last Run Time",type="date",JSONPath=".status.lastCreateRun"
// +kubebuilder:printcolumn:name="State",type="string",JSONPath=".status.state"

// SapNfsVolumeSnapshotSchedule is the Schema for the sapnfsvolumesnapshotschedules API
type SapNfsVolumeSnapshotSchedule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SapNfsVolumeSnapshotScheduleSpec   `json:"spec,omitempty"`
	Status SapNfsVolumeSnapshotScheduleStatus `json:"status,omitempty"`
}

func (sc *SapNfsVolumeSnapshotSchedule) Conditions() *[]metav1.Condition {
	return &sc.Status.Conditions
}

func (sc *SapNfsVolumeSnapshotSchedule) GetObjectMeta() *metav1.ObjectMeta {
	return &sc.ObjectMeta
}

func (sc *SapNfsVolumeSnapshotSchedule) SpecificToFeature() featuretypes.FeatureName {
	return featuretypes.FeatureNfsBackup
}

func (sc *SapNfsVolumeSnapshotSchedule) SpecificToProviders() []string {
	return []string{"openstack"}
}

func (sc *SapNfsVolumeSnapshotSchedule) State() string {
	return sc.Status.State
}

func (sc *SapNfsVolumeSnapshotSchedule) SetState(state string) {
	sc.Status.State = state
}

func (sc *SapNfsVolumeSnapshotSchedule) GetSourceRef() corev1.ObjectReference {
	return sc.Spec.Template.Spec.SourceVolume
}

func (sc *SapNfsVolumeSnapshotSchedule) SetSourceRef(ref corev1.ObjectReference) {
	sc.Spec.Template.Spec.SourceVolume = ref
}

func (sc *SapNfsVolumeSnapshotSchedule) GetSchedule() string {
	return sc.Spec.Schedule
}

func (sc *SapNfsVolumeSnapshotSchedule) SetSchedule(schedule string) {
	sc.Spec.Schedule = schedule
}

func (sc *SapNfsVolumeSnapshotSchedule) GetPrefix() string {
	return sc.Spec.Prefix
}

func (sc *SapNfsVolumeSnapshotSchedule) SetPrefix(prefix string) {
	sc.Spec.Prefix = prefix
}

func (sc *SapNfsVolumeSnapshotSchedule) GetStartTime() *metav1.Time {
	return sc.Spec.StartTime
}

func (sc *SapNfsVolumeSnapshotSchedule) SetStartTime(start *metav1.Time) {
	sc.Spec.StartTime = start
}

func (sc *SapNfsVolumeSnapshotSchedule) GetEndTime() *metav1.Time {
	return sc.Spec.EndTime
}

func (sc *SapNfsVolumeSnapshotSchedule) SetEndTime(end *metav1.Time) {
	sc.Spec.EndTime = end
}

func (sc *SapNfsVolumeSnapshotSchedule) GetMaxRetentionDays() int {
	return sc.Spec.MaxRetentionDays
}

func (sc *SapNfsVolumeSnapshotSchedule) SetMaxRetentionDays(days int) {
	sc.Spec.MaxRetentionDays = days
}

func (sc *SapNfsVolumeSnapshotSchedule) GetSuspend() bool {
	return sc.Spec.Suspend
}

func (sc *SapNfsVolumeSnapshotSchedule) SetSuspend(suspend bool) {
	sc.Spec.Suspend = suspend
}

func (sc *SapNfsVolumeSnapshotSchedule) GetDeleteCascade() bool {
	return sc.Spec.DeleteCascade
}

func (sc *SapNfsVolumeSnapshotSchedule) SetDeleteCascade(cascade bool) {
	sc.Spec.DeleteCascade = cascade
}

func (sc *SapNfsVolumeSnapshotSchedule) GetMaxReadyBackups() int {
	return sc.Spec.MaxReadySnapshots
}

func (sc *SapNfsVolumeSnapshotSchedule) SetMaxReadyBackups(count int) {
	sc.Spec.MaxReadySnapshots = count
}

func (sc *SapNfsVolumeSnapshotSchedule) GetMaxFailedBackups() int {
	return sc.Spec.MaxFailedSnapshots
}

func (sc *SapNfsVolumeSnapshotSchedule) SetMaxFailedBackups(count int) {
	sc.Spec.MaxFailedSnapshots = count
}

func (sc *SapNfsVolumeSnapshotSchedule) GetNextRunTimes() []string {
	return sc.Status.NextRunTimes
}

func (sc *SapNfsVolumeSnapshotSchedule) SetNextRunTimes(times []string) {
	sc.Status.NextRunTimes = times
}

func (sc *SapNfsVolumeSnapshotSchedule) GetNextDeleteTimes() map[string]string {
	return sc.Status.NextDeleteTimes
}

func (sc *SapNfsVolumeSnapshotSchedule) SetNextDeleteTimes(times map[string]string) {
	sc.Status.NextDeleteTimes = times
}

func (sc *SapNfsVolumeSnapshotSchedule) GetLastCreateRun() *metav1.Time {
	return sc.Status.LastCreateRun
}

func (sc *SapNfsVolumeSnapshotSchedule) SetLastCreateRun(time *metav1.Time) {
	sc.Status.LastCreateRun = time
}

func (sc *SapNfsVolumeSnapshotSchedule) GetLastCreatedBackup() corev1.ObjectReference {
	return sc.Status.LastCreatedBackup
}

func (sc *SapNfsVolumeSnapshotSchedule) SetLastCreatedBackup(obj corev1.ObjectReference) {
	sc.Status.LastCreatedBackup = obj
}

func (sc *SapNfsVolumeSnapshotSchedule) GetLastDeleteRun() *metav1.Time {
	return sc.Status.LastDeleteRun
}

func (sc *SapNfsVolumeSnapshotSchedule) SetLastDeleteRun(time *metav1.Time) {
	sc.Status.LastDeleteRun = time
}

func (sc *SapNfsVolumeSnapshotSchedule) GetLastDeletedBackups() []corev1.ObjectReference {
	return sc.Status.LastDeletedBackups
}

func (sc *SapNfsVolumeSnapshotSchedule) SetLastDeletedBackups(objs []corev1.ObjectReference) {
	sc.Status.LastDeletedBackups = objs
}

func (sc *SapNfsVolumeSnapshotSchedule) GetActiveSchedule() string {
	return sc.Status.Schedule
}

func (sc *SapNfsVolumeSnapshotSchedule) SetActiveSchedule(schedule string) {
	sc.Status.Schedule = schedule
}

func (sc *SapNfsVolumeSnapshotSchedule) GetBackupIndex() int {
	return sc.Status.SnapshotIndex
}

func (sc *SapNfsVolumeSnapshotSchedule) SetBackupIndex(index int) {
	sc.Status.SnapshotIndex = index
}

func (sc *SapNfsVolumeSnapshotSchedule) GetBackupCount() int {
	return sc.Status.BackupCount
}

func (sc *SapNfsVolumeSnapshotSchedule) SetBackupCount(count int) {
	sc.Status.BackupCount = count
}

func (sc *SapNfsVolumeSnapshotSchedule) CloneForPatchStatus() client.Object {
	return &SapNfsVolumeSnapshotSchedule{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SapNfsVolumeSnapshotSchedule",
			APIVersion: GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: sc.Namespace,
			Name:      sc.Name,
		},
		Status: sc.Status,
	}
}

// +kubebuilder:object:root=true

// SapNfsVolumeSnapshotScheduleList contains a list of SapNfsVolumeSnapshotSchedule
type SapNfsVolumeSnapshotScheduleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SapNfsVolumeSnapshotSchedule `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SapNfsVolumeSnapshotSchedule{}, &SapNfsVolumeSnapshotScheduleList{})
}
