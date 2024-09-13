package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusInitiated(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	if state.ObjAsVpcPeering().Status.State == "" {
		state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateInitiated
		changed = true
	}

	if !changed {
		return nil, ctx
	}

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		ErrorLogMessage("Error setting KCP VpcPeering Creating status state").
		SuccessErrorNil().
		Run(ctx, state)
}
