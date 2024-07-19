package backupschedule

import (
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type BackupSchedule interface {
	composed.ObjWithConditionsAndState
	BackupScheduleSpec
	BackupScheduleStatus
	GetList() client.ObjectList
}

type BackupScheduleSpec interface {
	GetSourceRef() corev1.ObjectReference
	SetSourceRef(ref corev1.ObjectReference)
	GetSchedule() string
	SetSchedule(schedule string)
	GetPrefix() string
	SetPrefix(prefix string)
	GetStartTime() *metav1.Time
	SetStartTime(start *metav1.Time)
	GetEndTime() *metav1.Time
	SetEndTime(end *metav1.Time)
	GetMaxRetentionDays() int
	SetMaxRetentionDays(days int)
	GetSuspend() bool
	SetSuspend(suspend bool)
	GetDeleteCascade() bool
	SetDeleteCascade(cascade bool)
}

type BackupScheduleStatus interface {
	GetNextRunTimes() []string
	SetNextRunTimes(times []string)
	GetNextDeleteTimes() map[string]string
	SetNextDeleteTimes(times map[string]string)
	GetLastCreateRun() *metav1.Time
	SetLastCreateRun(time *metav1.Time)
	GetLastCreatedBackup() corev1.ObjectReference
	SetLastCreatedBackup(obj corev1.ObjectReference)
	GetLastDeleteRun() *metav1.Time
	SetLastDeleteRun(time *metav1.Time)
	GetLastDeletedBackups() []corev1.ObjectReference
	SetLastDeletedBackups(objs []corev1.ObjectReference)
	GetActiveSchedule() string
	SetActiveSchedule(schedule string)
	GetBackupIndex() int
	SetBackupIndex(index int)
}
