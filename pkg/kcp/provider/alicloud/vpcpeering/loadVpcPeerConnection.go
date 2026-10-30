package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	alicloudvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/vpcpeering/client"
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
	for _, item := range list {
		if alicloudvpcpeeringclient.IsTerminal(item.Status) {
			continue
		}
		// Re-get by id so state holds the full GetVpcPeerConnectionAttribute body (the list
		// projection omits AcceptingVpc/CIDRs that downstream route actions need).
		info, err := state.client.GetVpcPeerConnection(ctx, item.InstanceId)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error loading AliCloud VpcPeerConnection found by name", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
		}
		if info != nil {
			state.vpcPeering = info
			if obj.Status.Id == "" {
				obj.Status.Id = item.InstanceId
				return composed.PatchStatus(obj).
					ErrorLogMessage("Error patching AliCloud VpcPeering status with recovered connection ID").
					FailedError(composed.StopWithRequeue).
					SuccessErrorNil().
					Run(ctx, state)
			}
			break
		}
	}

	return nil, ctx
}
