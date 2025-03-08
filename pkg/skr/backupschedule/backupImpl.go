package backupschedule

import (
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ScheduleType int

const (
	AwsNfsBackupSchedule ScheduleType = iota
	GcpNfsBackupSchedule
	AzureRwxBackupSchedule
)

type backupImpl interface {
	emptyScheduleObject() composed.ObjWithConditionsAndState
	emptySourceObject() client.Object
	emptyBackupList() client.ObjectList
	toObjectSlice(list client.ObjectList) []client.Object
	getBackupObject(state *State, objectMeta *metav1.ObjectMeta) (client.Object, error)
	sourceToObjWithConditionAndState(obj client.Object) (composed.ObjWithConditionsAndState, error)
}

func getBackupImpl(scheduleType ScheduleType) backupImpl {
	switch scheduleType {
	case AwsNfsBackupSchedule:
		return &backupImplAwsNfs{}
	case GcpNfsBackupSchedule:
		return &backupImplGcpNfs{}
	case AzureRwxBackupSchedule:
		return &backupImplAzureRwx{}
	default:
		return nil
	}
}
