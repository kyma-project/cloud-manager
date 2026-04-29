package sapnfsvolumesnapshot

import (
	"context"
	"fmt"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func snapshotWaitDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	// If snapshot was not loaded, it's already gone
	if state.snapshot == nil {
		return nil, ctx
	}

	// Re-fetch current status
	manilaSnapshot, err := state.snapshotClient.GetSnapshot(ctx, snapshot.Status.OpenstackId)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error checking Manila snapshot deletion status", composed.StopWithRequeue, ctx)
	}

	// Not found means deleted
	if manilaSnapshot == nil {
		return nil, ctx
	}

	switch manilaSnapshot.Status {
	case "deleting":
		return composed.StopWithRequeueDelay(10 * time.Second), ctx

	case "error_deleting":
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(fmt.Errorf("manila snapshot deletion failed"), "Manila snapshot is in error_deleting state", "openstackId", snapshot.Status.OpenstackId)
		snapshot.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(snapshot).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: "Manila snapshot deletion failed (error_deleting)",
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)

	default:
		return composed.StopWithRequeueDelay(10 * time.Second), ctx
	}
}
