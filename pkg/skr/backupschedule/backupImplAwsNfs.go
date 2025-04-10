package backupschedule

import (
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type backupImplAwsNfs struct {
}

func (impl *backupImplAwsNfs) emptyScheduleObject() composed.ObjWithConditionsAndState {
	return &cloudresourcesv1beta1.AwsNfsBackupSchedule{}
}

func (impl *backupImplAwsNfs) emptySourceObject() client.Object {
	return &cloudresourcesv1beta1.AwsNfsVolume{}
}

func (impl *backupImplAwsNfs) emptyBackupList() client.ObjectList {
	return &cloudresourcesv1beta1.AwsNfsVolumeBackupList{}
}
func (impl *backupImplAwsNfs) toObjectSlice(list client.ObjectList) []client.Object {
	var objects []client.Object

	if x, ok := list.(*cloudresourcesv1beta1.AwsNfsVolumeBackupList); ok {
		for _, item := range x.Items {
			objects = append(objects, &item)
		}
	}
	return objects
}
func (impl *backupImplAwsNfs) getBackupObject(state *State, objectMeta *metav1.ObjectMeta) (client.Object, error) {
	schedule := state.ObjAsBackupSchedule()
	_, ok := schedule.(*cloudresourcesv1beta1.AwsNfsBackupSchedule)
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", state.Scope.Spec.Provider)
	}

	return &cloudresourcesv1beta1.AwsNfsVolumeBackup{
		ObjectMeta: *objectMeta,
		Spec: cloudresourcesv1beta1.AwsNfsVolumeBackupSpec{
			Source: cloudresourcesv1beta1.AwsNfsVolumeBackupSource{
				Volume: cloudresourcesv1beta1.VolumeRef{
					Name:      schedule.GetSourceRef().Name,
					Namespace: schedule.GetSourceRef().Namespace,
				},
			},
		},
	}, nil
}

func (impl *backupImplAwsNfs) sourceToObjWithConditionAndState(obj client.Object) (composed.ObjWithConditionsAndState, error) {
	x, ok := obj.(*cloudresourcesv1beta1.AwsNfsVolume)
	if !ok {
		return nil, errors.New("source Object should be of type AwsNfsVolume")
	}
	return x, nil
}
