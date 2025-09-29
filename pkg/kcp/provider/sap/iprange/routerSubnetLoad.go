package iprange

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func routerSubnetLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.router == nil {
		return fmt.Errorf("router is required for SAP KCP IpRange routerSubnetLoad: %w", common.ErrLogical), ctx
	}
	if state.subnet == nil {
		return nil, ctx
	}

	arr, err := state.sapClient.ListRouterSubnetInterfaces(ctx, state.router.ID)
	if err != nil {
		logger.Error(err, "Error listing router subnets for SAP KCP IpRange")
		return composed.StopWithRequeue, ctx
	}

	for _, iface := range arr {
		if iface.SubnetID == state.subnet.ID {
			ii := iface
			state.routerSubnet = &ii
			break
		}
	}

	return nil, ctx
}
