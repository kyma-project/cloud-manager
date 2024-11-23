package vpcpeering

import (
	"context"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusInitiated(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	obj := state.ObjAsVpcPeering()
	if obj.Status.State == "" {
		state.ObjAsVpcPeering().Status.State = cloudcontrol1beta1.VpcPeeringConnectionStateReasonCodeInitiatingRequest
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
