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
	// Filter out terminal entries so we don't block on a Deleting/Deleted connection found by name.
	for i := range list {
		if list[i].Status != "Deleted" && list[i].Status != "Deleting" {
			state.vpcPeering = &list[i]
			break
		}
	}
	// If the list lookup recovered a connection that Status.Id doesn't know about yet, persist the
	// instance ID now. Without this, waitVpcPeeringActive would call GetVpcPeerConnection("").
	if state.vpcPeering != nil && obj.Status.Id != state.vpcPeering.InstanceId {
		obj.Status.Id = state.vpcPeering.InstanceId
		return composed.PatchStatus(obj).
			ErrorLogMessage("Error patching AliCloud VpcPeering status with recovered connection ID").
			FailedError(composed.StopWithRequeue).
			SuccessErrorNil().
			Run(ctx, state)
	}

	return nil, ctx
}
