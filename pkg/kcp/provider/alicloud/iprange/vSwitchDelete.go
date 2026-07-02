package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func vSwitchDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vSwitch == nil {
		return nil, ctx
	}

	logger.Info("Deleting AliCloud VSwitch for IpRange", "vSwitchId", state.vSwitch.VSwitchId)

	err := state.client.DeleteVSwitch(ctx, state.vSwitch.VSwitchId)
	if err != nil {
		logger.Error(err, "Error deleting AliCloud VSwitch for IpRange")
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx
}
