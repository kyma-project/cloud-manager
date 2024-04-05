package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
)

func createVpcPeeringConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	con, err := state.client.CreateVpcPeeringConnection(ctx, state.vpc.VpcId, state.remoteVpc.VpcId, pointer.String(state.remoteRegion), state.remoteVpc.OwnerId)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating AWS VPC Peering Connection", composed.StopWithRequeue, ctx)
	}

	state.vpcPeeringConnection = con

	state.ObjAsVpcPeering().Status.ConnectionId = pointer.StringDeref(con.VpcPeeringConnectionId, "")

	logger = logger.WithValues("connectionId", pointer.StringDeref(state.vpcPeeringConnection.VpcPeeringConnectionId, ""))

	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
