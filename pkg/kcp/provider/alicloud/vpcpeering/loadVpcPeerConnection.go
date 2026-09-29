package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadVpcPeerConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.vpcPeering != nil {
		return nil, ctx
	}

	obj := state.ObjAsVpcPeering()

	if obj.Status.Id != "" {
		info, err := state.client.GetVpcPeerConnection(ctx, obj.Status.Id)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error loading AliCloud VpcPeerConnection by ID", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		if info != nil {
			state.vpcPeering = info
			return nil, ctx
		}
	}

	if state.localVpcId == "" {
		return nil, ctx
	}

	list, err := state.client.ListVpcPeerConnections(ctx, state.localVpcId, obj.GetName())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AliCloud VpcPeerConnections", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	if len(list) > 0 {
		state.vpcPeering = &list[0]
	}

	return nil, ctx
}
