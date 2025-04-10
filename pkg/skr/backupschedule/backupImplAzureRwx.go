package backupschedule

import (
	"errors"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type backupImplAzureRwx struct {
}

type azureRwxSource struct {
	*corev1.PersistentVolumeClaim
}

func (impl *backupImplAzureRwx) emptyScheduleObject() composed.ObjWithConditionsAndState {
	return &cloudresourcesv1beta1.AzureRwxBackupSchedule{}
}

func (impl *backupImplAzureRwx) emptySourceObject() client.Object {
	return &corev1.PersistentVolumeClaim{}
}

func (impl *backupImplAzureRwx) emptyBackupList() client.ObjectList {
	return &cloudresourcesv1beta1.AzureRwxVolumeBackupList{}
}
func (impl *backupImplAzureRwx) toObjectSlice(list client.ObjectList) []client.Object {
	var objects []client.Object

	if x, ok := list.(*cloudresourcesv1beta1.AzureRwxVolumeBackupList); ok {
		for _, item := range x.Items {
			objects = append(objects, &item)
		}
	}
	return objects
}
func (impl *backupImplAzureRwx) getBackupObject(state *State, objectMeta *metav1.ObjectMeta) (client.Object, error) {
	schedule := state.ObjAsBackupSchedule()
	x, ok := schedule.(*cloudresourcesv1beta1.AzureRwxBackupSchedule)
	if !ok {
		return nil, fmt.Errorf("provider %s not supported", state.Scope.Spec.Provider)
	}

	return &cloudresourcesv1beta1.AzureRwxVolumeBackup{
		ObjectMeta: *objectMeta,
		Spec: cloudresourcesv1beta1.AzureRwxVolumeBackupSpec{
			Location: x.Spec.Location,
			Source: cloudresourcesv1beta1.PvcSource{
				Pvc: cloudresourcesv1beta1.PvcRef{
					Name:      schedule.GetSourceRef().Name,
					Namespace: schedule.GetSourceRef().Namespace,
				},
			},
		},
	}, nil
}

func (impl *backupImplAzureRwx) sourceToObjWithConditionAndState(obj client.Object) (composed.ObjWithConditionsAndState, error) {
	x, ok := obj.(*corev1.PersistentVolumeClaim)
	if !ok {
		return nil, errors.New("source Object should be of type PersistentVolumeClaim")
	}
	return &azureRwxSource{x}, nil
}

func (src *azureRwxSource) Conditions() *[]metav1.Condition {
	var conditions []metav1.Condition
	state := src.State()
	if state != "" {
		conditions = append(conditions, metav1.Condition{
			Type:    state,
			Status:  metav1.ConditionTrue,
			Reason:  state,
			Message: fmt.Sprintf("Pvc phase : %s", state),
		})
	}
	return &conditions
}
func (src *azureRwxSource) GetObjectMeta() *metav1.ObjectMeta {
	return &src.ObjectMeta
}

func (src *azureRwxSource) State() string {
	switch src.Status.Phase {
	case corev1.ClaimPending:
		return cloudresourcesv1beta1.StateProcessing
	case corev1.ClaimBound:
		return cloudresourcesv1beta1.StateReady
	case corev1.ClaimLost:
		return cloudresourcesv1beta1.StateError
	default:
		return ""
	}

}
func (src *azureRwxSource) SetState(_ string) {
	//NOOP
}
