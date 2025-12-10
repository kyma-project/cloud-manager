package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// createOrUpdatePsaConnection creates or updates the PSA connection based on whether it exists.
// This is a convenience action that routes to the appropriate create/update action.
func createOrUpdatePsaConnection(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	// If PSA connection doesn't exist, create it
	if state.serviceConnection == nil {
		return createPsaConnection(ctx, st)
	}

	// If PSA connection exists, update it
	return updatePsaConnection(ctx, st)
}
