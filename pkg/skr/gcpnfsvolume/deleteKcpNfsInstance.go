package gcpnfsvolume

import (
	"context"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func deleteKcpNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR GcpNfsVolume is NOT marked for deletion, do not delete mirror in KCP
		return nil, ctx
	}

	if state.KcpNfsInstance == nil {
		// SKR GcpNfsVolume is marked for deletion, but none found in KCP, probably already deleted
		return nil, ctx
	}
	if composed.IsMarkedForDeletion(state.KcpNfsInstance) {
		// KCP GcpNfsVolume is already marked for deletion, move forward to update the status in SKR if deletion failed
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Deleting KcpNfsInstance")

	err := state.KcpCluster.K8sClient().Delete(ctx, state.KcpNfsInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP NfsInstance", composed.StopWithRequeue, ctx)
	}

	// give some time to cloud-control and cloud providers to delete it, and then run again
	return composed.StopWithRequeueDelay(3 * time.Second), ctx
}
