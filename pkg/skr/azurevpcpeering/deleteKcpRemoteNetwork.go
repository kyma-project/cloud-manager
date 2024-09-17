package azurevpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteRemoteNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		// SKR AzureVpcPeering is NOT marked for deletion, do not delete mirror KCP
		return nil, nil
	}

	if state.RemoteNetwork == nil {
		// SKR RemoteNetwork is marked for deletion, but none found in KCP, probably already deleted
		return nil, nil
	}

	if composed.IsMarkedForDeletion(state.RemoteNetwork) {
		return nil, nil
	}

	logger.Info("Deleting KCP remote network")

	err := state.KcpCluster.K8sClient().Delete(ctx, state.RemoteNetwork)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP remote Network", composed.StopWithRequeue, ctx)
	}

	// give some time to cloud-control and cloud providers to delete it, and then run again
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
