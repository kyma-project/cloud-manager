package subnet

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func addSubnetToConnectionPolicy(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.serviceConnectionPolicy == nil {
		return composed.StopWithRequeue, nil
	}
	if state.subnet == nil {
		return composed.StopWithRequeue, nil
	}
	if state.ConnectionPolicySubnetsContainCurrent() {
		return nil, nil
	}

	state.AddCurrentSubnetToConnectionPolicy()

	return nil, nil
}
