package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func copyCidrToStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if len(state.ObjAsIpRange().Status.Cidr) > 0 ||
		state.ObjAsIpRange().Status.Cidr == state.ObjAsIpRange().Spec.Cidr {
		return nil, ctx
	}

	state.ObjAsIpRange().Status.Cidr = state.ObjAsIpRange().Spec.Cidr

	return composed.PatchStatus(state.ObjAsIpRange()).
		SuccessErrorNil().
		Run(ctx, state)
}
