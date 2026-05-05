package sapnfsvolumesnapshotrestore

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func restoreInPlaceWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsSapNfsVolumeSnapshotRestore()

	share, err := state.shareClient.GetShare(ctx, state.shareId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting share status during revert", composed.StopWithRequeue, ctx)
	}

	state.share = share

	switch share.Status {
	case "available":
		// Revert completed — set Done and continue to releaseLease
		restore.Status.State = cloudresourcesv1beta1.JobStateDone
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonReady,
				Message: "In-place restore completed successfully",
			}).
			OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
				return nil, ctx
			}).
			Run(ctx, state)

	case "reverting":
		// Still reverting, requeue after delay
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx

	case "reverting_error":
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(fmt.Errorf("share revert failed with status reverting_error"), "Share revert failed")
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonNfsRestoreFailed,
				Message: "Share revert operation failed with status reverting_error",
			}).
			OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
				return nil, ctx
			}).
			Run(ctx, state)

	default:
		// Unknown status, requeue
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}
}
