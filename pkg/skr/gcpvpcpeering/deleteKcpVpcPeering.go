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

	if state.KcpVpcPeering == nil {
		return nil, nil
	}

	if composed.IsMarkedForDeletion(state.KcpVpcPeering) {
		logger.Info("KCP VpcPeering is marked for deletion, re-queueing until it is deleted.", "kcpVpcPeering", state.KcpVpcPeering.Name)
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	logger.Info("Deleting KCP VpcPeering", "kcpVpcPeering", state.KcpVpcPeering.Name)

	err := state.KcpCluster.K8sClient().Delete(ctx, state.KcpVpcPeering)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP VpcPeering ", composed.StopWithRequeue, ctx)
	}

	state.ObjAsGcpVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDeleting
	return composed.PatchStatus(state.ObjAsGcpVpcPeering()).
		ErrorLogMessage("Error patching status").
		FailedError(composed.StopWithRequeue).
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
		Run(ctx, state)
}
