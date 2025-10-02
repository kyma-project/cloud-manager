package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func subnetLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	subnet, err := state.sapClient.GetSubnetByName(ctx, state.net.ID, state.SubnetName())
	if err != nil {
		logger.Error(err, "Error loading Openstack subnet for IpRange")
		return composed.StopWithRequeue, ctx
	}

	if subnet != nil {
		state.subnet = subnet
	}

	return nil, ctx
}
