package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func vSwitchDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if len(state.vSwitches) == 0 {
		return nil, ctx
	}

	for _, vsw := range state.vSwitches {
		logger.Info("Deleting AliCloud VSwitch for IpRange", "vSwitchId", vsw.VSwitchId)

		err := state.client.DeleteVSwitch(ctx, vsw.VSwitchId)
		if err != nil {
			logger.Error(err, "Error deleting AliCloud VSwitch for IpRange", "vSwitchId", vsw.VSwitchId)
			return composed.StopWithRequeue, ctx
		}
	}

	// Requeue to let AliCloud finish removing vSwitches before attempting
	// to disassociate the VPC address space (which fails if vSwitches still exist).
	return composed.StopWithRequeue, ctx
}
