package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func vSwitchLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Load by IDs already in status
	if len(state.ObjAsIpRange().Status.Subnets) > 0 {
		for _, subnet := range state.ObjAsIpRange().Status.Subnets {
			if subnet.Id == "" {
				continue
			}
			vsw, err := state.client.DescribeVSwitch(ctx, subnet.Id)
			if err != nil {
				logger.Error(err, "Error loading AliCloud VSwitch by ID for IpRange", "vSwitchId", subnet.Id)
				return composed.StopWithRequeue, ctx
			}
			if vsw != nil {
				state.vSwitches = append(state.vSwitches, vsw)
			}
		}
		return nil, ctx
	}

	// Otherwise search by name prefix across all zones
	for i := range state.Scope().Spec.Scope.Alicloud.Network.Zones {
		name := state.VSwitchName(i)
		vswitches, err := state.client.DescribeVSwitchesByName(ctx, state.vpcId, name)
		if err != nil {
			logger.Error(err, "Error loading AliCloud VSwitch by name for IpRange", "name", name)
			return composed.StopWithRequeue, ctx
		}
		if len(vswitches) > 0 {
			state.vSwitches = append(state.vSwitches, &vswitches[0])
		}
	}

	return nil, ctx
}
