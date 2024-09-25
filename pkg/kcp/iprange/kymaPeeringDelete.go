package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func kymaPeeringDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	if state.kymaPeering == nil {
		return nil, ctx
	}

	if composed.IsMarkedForDeletion(state.kymaPeering) {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Deleting KCP Kyma VpcPeering for IpRange")

	err := state.Cluster().K8sClient().Delete(ctx, state.kymaPeering)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP Kyma VpcPeering for IpRange", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
