package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitVpcPeeringActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	obj := state.ObjAsVpcPeering()

	info, err := state.client.GetVpcPeerConnection(ctx, obj.Status.Id)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting AliCloud VpcPeerConnection status", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	if info == nil {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	state.vpcPeering = info

	switch info.Status {
	case "Activated":
		logger.Info("AliCloud VpcPeerConnection is Activated")
		return nil, ctx

	case "Accepting":
		logger.Info("AliCloud VpcPeerConnection is Accepting, calling AcceptVpcPeerConnection")
		if err := state.remoteClient.AcceptVpcPeerConnection(ctx, obj.Status.Id); err != nil {
			return composed.LogErrorAndReturn(err, "Error accepting AliCloud VpcPeerConnection", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx

	case "Deleted", "Deleting":
		logger.Info("AliCloud VpcPeerConnection is being deleted, resetting state")
		obj.Status.Id = ""
		return composed.PatchStatus(obj).
			ErrorLogMessage("Error patching AliCloud VpcPeering status after connection deleted").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)

	default:
		logger.Info("AliCloud VpcPeerConnection waiting", "status", info.Status)
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}
}
