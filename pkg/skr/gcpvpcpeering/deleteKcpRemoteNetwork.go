package gcpvpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteKcpRemoteNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.KcpRemoteNetwork == nil {
		return nil, nil
	}

	if composed.IsMarkedForDeletion(state.KcpRemoteNetwork) {
		return nil, nil
	}

	if state.KcpVpcPeering != nil {
		logger.Info("[SKR GCP VPCPeering deleteKcpRemoteNetwork] Waiting for KCP VpcPeering deletion")
		return composed.StopWithRequeueDelay(3 * util.Timing.T1000ms()), nil
	}

	logger.Info("[SKR GCP VPCPeering deleteKcpRemoteNetwork] Deleting GCP KCP Remote Network")

	err := state.KcpCluster.K8sClient().Delete(ctx, state.KcpRemoteNetwork)

	if err != nil {
		return composed.LogErrorAndReturn(err, "[SKR GCP VPCPeering deleteKcpRemoteNetwork] Error deleting GCP KCP remote Network", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
