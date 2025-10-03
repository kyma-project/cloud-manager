package iprange

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusPatch(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	if state.ObjAsIpRange().Status.VpcId != state.net.ID {
		state.ObjAsIpRange().Status.VpcId = state.net.ID
		changed = true
	}

	if len(state.ObjAsIpRange().Status.Subnets) != 1 {
		state.ObjAsIpRange().Status.Subnets = cloudcontrolv1beta1.IpRangeSubnets{{}}
		changed = true
	}
	if state.ObjAsIpRange().Status.Subnets[0].Id != state.subnet.ID {
		state.ObjAsIpRange().Status.Subnets[0].Id = state.subnet.ID
		changed = true
	}
	if state.ObjAsIpRange().Status.Subnets[0].Name != state.subnet.Name {
		state.ObjAsIpRange().Status.Subnets[0].Name = state.subnet.Name
		changed = true
	}
	if state.ObjAsIpRange().Status.Subnets[0].Range != state.subnet.CIDR {
		state.ObjAsIpRange().Status.Subnets[0].Range = state.subnet.CIDR
		changed = true
	}

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsIpRange()).
		ErrorLogMessage("Error patching SAP KCP IpRange for Openstack with subnet details").
		FailedError(composed.StopWithRequeue).
		SuccessErrorNil().
		Run(ctx, state)
}
