package iprange

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func routerSubnetAdd(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.routerSubnet != nil {
		return nil, ctx
	}
	if state.subnet == nil {
		return fmt.Errorf("subnet is required for SAP KCP IpRange routerSubnetAdd: %w", common.ErrLogical), ctx
	}

	_, err := state.sapClient.AddSubnetToRouter(ctx, state.router.ID, state.subnet.ID)
	if err != nil {
		logger.Error(err, "Error adding subnet to router for SAP KCP IpRange")
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx
}
