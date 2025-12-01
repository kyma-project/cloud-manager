package gcpvpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteKcpRemoteNetwork(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpRemoteNetwork == nil {
		return nil, nil
	}

	if composed.IsMarkedForDeletion(state.KcpRemoteNetwork) {
		logger.Info("KCP Remote Network is marked for deletion, will continue.")
		return nil, nil
	}

	logger.Info("Deleting GCP KCP Remote Network ", "remoteNetwork", state.KcpRemoteNetwork.Name)

	err := state.KcpCluster.K8sClient().Delete(ctx, state.KcpRemoteNetwork)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting GCP KCP remote Network "+state.KcpRemoteNetwork.Name, composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
