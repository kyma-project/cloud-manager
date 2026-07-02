package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func vSwitchWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// If we already loaded a vSwitch and it's Available, proceed
	if state.vSwitch != nil && state.vSwitch.Status == "Available" {
		return nil, ctx
	}

	// Load the VSwitch to check its current status
	if state.vSwitchId == "" {
		logger.Info("AliCloud VSwitch ID is empty, requeueing")
		return composed.StopWithRequeue, ctx
	}

	vsw, err := state.client.DescribeVSwitch(ctx, state.vSwitchId)
	if err != nil {
		logger.Error(err, "Error describing AliCloud VSwitch while waiting")
		return composed.StopWithRequeue, ctx
	}

	if vsw == nil {
		logger.Info("AliCloud VSwitch not found while waiting, requeueing")
		return composed.StopWithRequeue, ctx
	}

	state.vSwitch = vsw

	if vsw.Status != "Available" {
		logger.Info("AliCloud VSwitch not yet available, requeueing", "status", vsw.Status)
		return composed.StopWithRequeue, ctx
	}

	return nil, ctx
}
