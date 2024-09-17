package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shouldPeerWithKymaNetwork(ctx context.Context, st composed.State) bool {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return false
	}

	// can not peer with itself
	if state.isCloudManagerNetwork {
		return false
	}

	if focal.AzureProviderPredicate(ctx, state) && state.isCloudManagerNetwork {
		return true
	}

	return false
}
