package vpcpeering

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func statusCreating(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	changed := false

	if state.ObjAsVpcPeering().Status.State == "" {
		state.ObjAsVpcPeering().Status.State = "Creating"
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
