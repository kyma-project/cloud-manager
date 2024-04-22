package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
)

func loadVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	connectionId := state.ObjAsVpcPeering().Status.ConnectionId

	// skip loading of vpc peering connections if connectionId is empty
	if len(connectionId) == 0 {
		return nil, ctx
	}

	list, err := state.client.DescribeVpcPeeringConnections(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AWS peering connections", composed.StopWithRequeue, ctx)
	}

	for _, c := range list {
		if connectionId == pointer.StringDeref(c.VpcPeeringConnectionId, "") {
			state.vpcPeeringConnection = &c
			break
		}
	}

	if state.vpcPeeringConnection == nil {
		return nil, ctx
	}

	logger = logger.WithValues(
		"vpcConnectionId", pointer.StringDeref(state.vpcPeeringConnection.VpcPeeringConnectionId, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
