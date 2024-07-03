package awsvpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteKcpVpcPeering(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.KcpVpcPeering == nil {
		// SKR AwsVpcPeering is marked for deletion, but none found in KCP, probably already deleted
		return nil, nil
	}

	if composed.IsMarkedForDeletion(state.KcpVpcPeering) {
		return nil, nil
	}

	logger.Info("Deleting KCP VpcPeering")

	err := state.KcpCluster.K8sClient().Delete(ctx, state.KcpVpcPeering)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating KCP VpcPeering", composed.StopWithRequeue, ctx)
	}

	// give some time to cloud-control and cloud providers to delete it, and then run again
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
