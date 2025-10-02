package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func subnetDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.subnet == nil {
		return nil, ctx
	}

	err := state.sapClient.DeleteSubnet(ctx, state.subnet.ID)
	if err != nil {
		logger.Error(err, "Error deleting Openstack subnet for SAP KCP IpRange")
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx
}
