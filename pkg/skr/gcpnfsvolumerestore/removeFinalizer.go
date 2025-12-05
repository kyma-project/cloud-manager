package gcpnfsvolumerestore

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
			SuccessError(composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime)).
			Run(ctx, state)
	}

	modified, err := st.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR GcpNfsVolumeRestore after finalizer remove", composed.StopWithRequeue, ctx)
	}
	if modified {
		composed.LoggerFromCtx(ctx).Info("Finalizer removed")
	}
	return composed.StopAndForget, nil
}
