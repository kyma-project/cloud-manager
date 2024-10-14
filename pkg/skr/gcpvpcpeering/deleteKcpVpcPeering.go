package gcpvpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
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
		// VpcPeering on SKR is marked for deletion, but not found in KCP, so probably it is already deleted
		return nil, nil
	}

	if composed.IsMarkedForDeletion(state.KcpVpcPeering) {
		return nil, nil
	}

	logger.Info("Deleting KCP VpcPeering")

	err := state.KcpCluster.K8sClient().Delete(ctx, state.KcpVpcPeering)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP VpcPeering", composed.StopWithRequeue, ctx)
	}

	state.ObjAsGcpVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDeleting
	return composed.PatchStatus(state.ObjAsGcpVpcPeering()).
		ErrorLogMessage("[SKR GCP VPCPeering deleteKcpVpcPeering] Error patching status").
		FailedError(composed.StopWithRequeue).
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
		Run(ctx, state)
}
