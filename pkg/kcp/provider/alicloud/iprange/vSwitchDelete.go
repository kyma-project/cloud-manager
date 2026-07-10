package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func vSwitchDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for _, vsw := range state.vSwitches {
		logger.Info("Deleting AliCloud VSwitch for IpRange", "vSwitchId", vsw.VSwitchId)

		err := state.client.DeleteVSwitch(ctx, vsw.VSwitchId)
		if err != nil {
			logger.Error(err, "Error deleting AliCloud VSwitch for IpRange", "vSwitchId", vsw.VSwitchId)
			return composed.StopWithRequeue, ctx
		}
	}

	return nil, ctx
}
