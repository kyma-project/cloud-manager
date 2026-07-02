package iprange

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusPatch(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.vSwitch == nil {
		return nil, ctx
	}

	changed := false

	if state.ObjAsIpRange().Status.VpcId != state.vpcId {
		state.ObjAsIpRange().Status.VpcId = state.vpcId
		changed = true
	}

	expectedSubnets := cloudcontrolv1beta1.IpRangeSubnets{{
		Id:    state.vSwitch.VSwitchId,
		Zone:  state.vSwitch.ZoneId,
		Range: state.vSwitch.CidrBlock,
		Name:  state.vSwitch.VSwitchName,
	}}

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
