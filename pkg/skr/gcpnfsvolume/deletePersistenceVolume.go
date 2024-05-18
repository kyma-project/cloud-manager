package gcpnfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func deletePersistenceVolume(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)

	//If GcpNfsVolume is not marked for deletion, continue
	if !composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR GcpNfsVolume is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}

	//If PV doesn't exist or already marked for Deletion, continue
	if state.PV == nil || !state.PV.DeletionTimestamp.IsZero() {
		return nil, nil
	}

	if state.PV.Status.Phase != "Released" && state.PV.Status.Phase != "Available" {
		// Only PV in Released or Available state can be deleted
		state.ObjAsGcpNfsVolume().Status.State = cloudresourcesv1beta1.GcpNfsVolumeError
		return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonPVNotReadyForDeletion,
				Message: fmt.Sprintf("Deletion failed because associated Persistent Volume %s cannot be deleted. Only Persistent Volumes in Released or Available state can be deleted.", state.PV.Name),
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error updating GcpNfsVolume status with Persistent Volume not ready for deletion").
			Run(ctx, state)
	} else {
		// Remove conditionType Error if it was the result of the improper PV status that was fixed
		for i := range state.ObjAsGcpNfsVolume().Status.Conditions {
			condition := (state.ObjAsGcpNfsVolume().Status.Conditions)[i]
			if condition.Type == cloudresourcesv1beta1.ConditionTypeError && condition.Reason == cloudresourcesv1beta1.ConditionReasonPVNotReadyForDeletion {
				return composed.UpdateStatus(state.ObjAsGcpNfsVolume()).
					RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
					ErrorLogMessage("Error removing conditionType Error").
					OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
						return composed.StopWithRequeue, nil
					}).
					Run(ctx, state)
			}
		}
	}

	//Delete PV
	err := state.SkrCluster.K8sClient().Delete(ctx, state.PV)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting PersistentVolume", composed.StopWithRequeue, ctx)
	}

	// give some time, and then run again
	return composed.StopWithRequeueDelay(3 * time.Second), nil
}
