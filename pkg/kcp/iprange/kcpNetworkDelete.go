package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func kcpNetworkDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	if state.network == nil {
		return nil, ctx
	}

	if composed.IsMarkedForDeletion(state.network) {
		return nil, ctx
	}

	if !state.isCloudManagerNetwork {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Deleting KCP CM Network for IpRange")

	err := state.Cluster().K8sClient().Delete(ctx, state.network)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP CM Network for IpRange", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
