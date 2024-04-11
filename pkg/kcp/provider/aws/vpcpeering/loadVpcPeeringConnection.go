package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
)

func loadVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	list, err := state.client.DescribeVpcPeeringConnections(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AWS peering connections", composed.StopWithRequeue, ctx)
	}

	// TODO use Status.ConnectionId
	for _, c := range list {
		if state.ObjAsVpcPeering().Status.ConnectionId == pointer.StringDeref(c.VpcPeeringConnectionId, "") {
			state.vpcPeeringConnection = &c
			break
		}
	}

	if state.vpcPeeringConnection == nil {
		return nil, nil
	}

	logger = logger.WithValues(
		"vpcConnectionId", pointer.StringDeref(state.vpcPeeringConnection.VpcPeeringConnectionId, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	if len(state.ObjAsVpcPeering().Status.ConnectionId) > 0 {
		return nil, ctx
	}

	state.ObjAsVpcPeering().Status.ConnectionId = *state.vpcPeeringConnection.VpcPeeringConnectionId

	err = state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating VPC Peering status with connection id", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, ctx
}
