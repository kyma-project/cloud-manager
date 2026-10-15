package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func acceptVpcPeerConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vpcPeering == nil || state.vpcPeering.Status != "Accepting" {
		return nil, ctx
	}

	obj := state.ObjAsVpcPeering()

	logger.Info("AliCloud VpcPeerConnection is Accepting, calling AcceptVpcPeerConnection")
	if err := state.remoteClient.AcceptVpcPeerConnection(ctx, obj.Status.Id); err != nil {
		return composed.LogErrorAndReturn(err, "Error accepting AliCloud VpcPeerConnection", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
