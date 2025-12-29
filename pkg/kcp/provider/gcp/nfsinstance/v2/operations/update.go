package operations

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// UpdateInstance updates an existing GCP Filestore instance.
// TODO: Implement this action.
func UpdateInstance(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Implementation placeholder
	// This will:
	// 1. Build update mask from state
	// 2. Call client.UpdateInstance()
	// 3. Store pending operation
	// 4. Set state to updating
	// 5. Return nil, ctx to continue
	return nil, ctx
}
