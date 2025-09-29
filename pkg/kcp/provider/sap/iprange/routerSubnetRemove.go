package iprange

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func routerSubnetRemove(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.router == nil {
		return fmt.Errorf("router is required for SAP KCP IpRange routerSubnetRemove: %w", common.ErrLogical), ctx
	}
	if state.routerSubnet == nil {
		return nil, ctx
	}
	if state.subnet == nil {
		return nil, ctx
	}

	err := state.sapClient.RemoveSubnetFromRouter(ctx, state.router.ID, state.subnet.ID)
	if err != nil {
		logger.Error(err, "Error removing subnet from router for SAP KCP IpRange")
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx
}
