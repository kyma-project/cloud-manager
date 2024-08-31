package gcpnfsvolumerestore

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsGcpNfsVolumeRestore()

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	// This shouldn't be reached, but just in case
	if restore.Status.State == "InProgress" {
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsRestoreInProgress,
				Message: "In progress restore cannot be deleted",
			}).
			SuccessError(composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime)).
			Run(ctx, state)
	}

	controllerutil.RemoveFinalizer(state.Obj(), cloudresourcesv1beta1.Finalizer)
	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR GcpNfsVolumeRestore after finalizer remove", composed.StopWithRequeue, ctx)
	}

	return composed.StopAndForget, nil
}
