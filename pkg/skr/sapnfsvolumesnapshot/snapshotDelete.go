package sapnfsvolumesnapshot

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func snapshotDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	// If snapshot not loaded (already gone), continue to finalizer removal
	if state.snapshot == nil {
		return nil, ctx
	}

	// If already in deleting status, skip to wait
	if state.snapshot.Status == "deleting" {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Deleting Manila snapshot", "openstackId", snapshot.Status.OpenstackId)

	err := state.snapshotClient.DeleteSnapshot(ctx, snapshot.Status.OpenstackId)
	if err != nil {
		logger.Error(err, "Error deleting Manila snapshot")
		snapshot.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(snapshot).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: fmt.Sprintf("Error deleting Manila snapshot: %s", err.Error()),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
