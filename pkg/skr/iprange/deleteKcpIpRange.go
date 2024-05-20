package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func deleteKcpIpRange(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR IpRange is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}

	if state.KcpIpRange == nil {
		// SKR IpRange is marked for deletion, but none found in KCP, probably already deleted
		return nil, nil
	}
	if composed.IsMarkedForDeletion(state.KcpIpRange) {
		// KCP IpRange is already marked for deletion, move forward to update the status in SKR if deletion failed
		return nil, nil
	}

	logger.Info("Deleting KCP IpRange")

	err := state.KcpCluster.K8sClient().Delete(ctx, state.KcpIpRange)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP IpRange", composed.StopWithRequeue, ctx)
	}
	state.ObjAsIpRange().SetState(v1beta1.StateDeleting)
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR IpRange status with Deleting state", composed.StopWithRequeue, ctx)
	}

	// give some time to cloud-control and cloud providers to delete it, and then run again
	return composed.StopWithRequeueDelay(300 * time.Millisecond), nil
}
