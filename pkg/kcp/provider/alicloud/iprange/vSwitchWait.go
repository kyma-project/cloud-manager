package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func vSwitchWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	for i, vsw := range state.vSwitches {
		if vsw.Status == "Available" {
			continue
		}

		// Reload to check current status
		updated, err := state.client.DescribeVSwitch(ctx, vsw.VSwitchId)
		if err != nil {
			logger.Error(err, "Error describing AliCloud VSwitch while waiting", "vSwitchId", vsw.VSwitchId)
			return composed.StopWithRequeue, ctx
		}
		if updated == nil {
			logger.Info("AliCloud VSwitch not found while waiting, requeueing", "vSwitchId", vsw.VSwitchId)
			return composed.StopWithRequeue, ctx
		}

		state.vSwitches[i] = updated

		if updated.Status != "Available" {
			logger.Info("AliCloud VSwitch not yet available, requeueing", "vSwitchId", vsw.VSwitchId, "status", updated.Status)
			return composed.StopWithRequeue, ctx
		}
	}

	return nil, ctx
}
