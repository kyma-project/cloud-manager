package iprange

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusPatch(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if len(state.vSwitches) == 0 {
		return nil, ctx
	}

	changed := false

	if state.ObjAsIpRange().Status.VpcId != state.vpcId {
		state.ObjAsIpRange().Status.VpcId = state.vpcId
		changed = true
	}

	expectedSubnets := make(cloudcontrolv1beta1.IpRangeSubnets, 0, len(state.vSwitches))
	for _, vsw := range state.vSwitches {
		expectedSubnets = append(expectedSubnets, cloudcontrolv1beta1.IpRangeSubnet{
			Id:    vsw.VSwitchId,
			Zone:  vsw.ZoneId,
			Range: vsw.CidrBlock,
			Name:  vsw.VSwitchName,
		})
	}

	if !state.ObjAsIpRange().Status.Subnets.Equals(expectedSubnets) {
		state.ObjAsIpRange().Status.Subnets = expectedSubnets
		changed = true
	}

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsIpRange()).
		ErrorLogMessage("Error patching AliCloud KCP IpRange with VSwitch details").
		FailedError(composed.StopWithRequeue).
		SuccessErrorNil().
		Run(ctx, state)
}
