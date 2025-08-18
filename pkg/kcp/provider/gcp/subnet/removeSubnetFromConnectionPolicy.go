package subnet

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func removeSubnetFromConnectionPolicy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.serviceConnectionPolicy == nil {
		return nil, ctx
	}

	if !state.ConnectionPolicySubnetsContainCurrent() {
		return nil, ctx
	}

	if state.ConnectionPolicySubnetsLen() == 1 { // last one in, cant remove
		return nil, ctx
	}

	state.RemoveCurrentSubnetFromConnectionPolicy()

	return nil, ctx
}
