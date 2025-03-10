package backupschedule

import (
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type backupImplGcpNfs struct {
}

func (impl *backupImplGcpNfs) emptyScheduleObject() composed.ObjWithConditionsAndState {
	return &cloudresourcesv1beta1.GcpNfsBackupSchedule{}
}

func (impl *backupImplGcpNfs) emptySourceObject() client.Object {
	return &cloudresourcesv1beta1.GcpNfsVolume{}
}

func (impl *backupImplGcpNfs) emptyBackupList() client.ObjectList {
	return &cloudresourcesv1beta1.GcpNfsVolumeBackupList{}
}
func (impl *backupImplGcpNfs) toObjectSlice(list client.ObjectList) []client.Object {
	var objects []client.Object

	if x, ok := list.(*cloudresourcesv1beta1.GcpNfsVolumeBackupList); ok {
		for _, item := range x.Items {
			objects = append(objects, &item)
		}
	}
	return objects
}
func (impl *backupImplGcpNfs) getBackupObject(state *State, objectMeta *metav1.ObjectMeta) (client.Object, error) {
	schedule := state.ObjAsBackupSchedule()
	x, ok := schedule.(*cloudresourcesv1beta1.GcpNfsBackupSchedule)
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", state.Scope.Spec.Provider)
	}

	return &cloudresourcesv1beta1.GcpNfsVolumeBackup{
		ObjectMeta: *objectMeta,
		Spec: cloudresourcesv1beta1.GcpNfsVolumeBackupSpec{
			Location: x.Spec.Location,
			Source: cloudresourcesv1beta1.GcpNfsVolumeBackupSource{
				Volume: cloudresourcesv1beta1.GcpNfsVolumeRef{
					Name:      schedule.GetSourceRef().Name,
					Namespace: schedule.GetSourceRef().Namespace,
				},
			},
		},
	}, nil
}

func (impl *backupImplGcpNfs) sourceToObjWithConditionAndState(obj client.Object) (composed.ObjWithConditionsAndState, error) {
	x, ok := obj.(*cloudresourcesv1beta1.GcpNfsVolume)
	if !ok {
		return nil, errors.New("Source Object should be of tyoe AwsNfsVolume")
	}
	return x, nil
}
