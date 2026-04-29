package sapnfsvolumesnapshot

import (
	"context"
	"fmt"
	"time"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func snapshotWaitAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	snapshot := state.ObjAsSapNfsVolumeSnapshot()

	if state.snapshot == nil {
		return nil, ctx
	}

	switch state.snapshot.Status {
	case "available":
		return nil, ctx

	case "creating":
		return composed.StopWithRequeueDelay(10 * time.Second), ctx

	case "error":
		logger := composed.LoggerFromCtx(ctx)
		logger.Error(fmt.Errorf("manila snapshot entered error state"), "Manila snapshot is in error state", "openstackId", snapshot.Status.OpenstackId)
		snapshot.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(snapshot).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: "Manila snapshot entered error state",
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)

	default:
		return composed.StopWithRequeueDelay(10 * time.Second), ctx
	}
}
