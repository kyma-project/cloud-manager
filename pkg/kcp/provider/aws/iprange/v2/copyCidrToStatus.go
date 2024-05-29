package v2

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func copyCidrToStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if len(state.ObjAsIpRange().Status.Cidr) > 0 {
		return nil, nil
	}

	state.ObjAsIpRange().Status.Cidr = state.ObjAsIpRange().Spec.Cidr

	return composed.PatchStatus(state.ObjAsIpRange()).
		SuccessErrorNil().
		Run(ctx, state)
}
