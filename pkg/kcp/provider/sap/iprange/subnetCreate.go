package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func subnetCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnet != nil {
		return nil, ctx
	}

	subnet, err := state.sapClient.CreateSubnetOp(ctx, state.net.ID, state.ObjAsIpRange().Status.Cidr, state.SubnetName())
	if err != nil {
		logger.Error(err, "Error creating Openstack subnet for SAP KCP IpRange")
		return composed.StopWithRequeue, ctx
	}

	state.subnet = subnet

	return nil, ctx
}
