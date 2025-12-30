package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/statewithscope"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shouldPeerWithKymaNetwork(ctx context.Context, st composed.State) bool {
	state := st.(*State)

	if composed.IsMarkedForDeletion(state.Obj()) {
		return false
	}

	// can not peer with itself
	if state.isKymaNetwork {
		return false
	}

	if statewithscope.AzureProviderPredicate(ctx, state) && state.isCloudManagerNetwork {
		return true
	}

	return false
}
