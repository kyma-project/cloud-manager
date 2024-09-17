package vpcpeering

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusInitiated(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	obj := state.ObjAsVpcPeering()
	if obj.Status.State == "" {
		state.ObjAsVpcPeering().Status.State = string(ec2Types.VpcPeeringConnectionStateReasonCodeInitiatingRequest)
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
