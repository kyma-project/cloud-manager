package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func setBandwidth(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	isCrossRegion := state.localRegion != state.remoteRegion
	if !isCrossRegion {
		return nil, ctx
	}
	if state.ObjAsVpcPeering().Spec.Details == nil {
		return nil, ctx
	}

	bandwidth := state.ObjAsVpcPeering().Spec.Details.Bandwidth
	if bandwidth == 0 {
		bandwidth = 1024
	}

	if state.vpcPeering != nil && state.vpcPeering.Bandwidth == bandwidth {
		return nil, ctx
	}

	instanceId := state.ObjAsVpcPeering().Status.Id
	if instanceId == "" {
		return nil, ctx
	}

	if err := state.client.ModifyVpcPeerConnection(ctx, instanceId, bandwidth); err != nil {
		if alicloudvpcpeeringclient.IsRetryable(err) {
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
		}
		return composed.LogErrorAndReturn(err, "Error setting AliCloud VpcPeerConnection bandwidth", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
}
