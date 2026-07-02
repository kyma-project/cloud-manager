package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func vSwitchLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// If we already have a VSwitch ID in status, load it directly
	if len(state.ObjAsIpRange().Status.Subnets) > 0 && state.ObjAsIpRange().Status.Subnets[0].Id != "" {
		vsw, err := state.client.DescribeVSwitch(ctx, state.ObjAsIpRange().Status.Subnets[0].Id)
		if err != nil {
			logger.Error(err, "Error loading AliCloud VSwitch by ID for IpRange")
			return composed.StopWithRequeue, ctx
		}
		if vsw != nil {
			state.vSwitch = vsw
			state.vSwitchId = vsw.VSwitchId
		}
		return nil, ctx
	}

	// Otherwise, search by name within the VPC
	vswitches, err := state.client.DescribeVSwitchesByName(ctx, state.vpcId, state.VSwitchName())
	if err != nil {
		logger.Error(err, "Error loading AliCloud VSwitch by name for IpRange")
		return composed.StopWithRequeue, ctx
	}

	if len(vswitches) > 0 {
		state.vSwitch = &vswitches[0]
		state.vSwitchId = vswitches[0].VSwitchId
	}

	return nil, ctx
}
