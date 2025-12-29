package operations

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// DeleteInstance deletes a GCP Filestore instance.
// TODO: Implement this action.
func DeleteInstance(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Implementation placeholder
	// This will:
	// 1. Call client.DeleteInstance()
	// 2. Store pending operation
	// 3. Set state to deleting
	// 4. Return nil, ctx to continue
	return nil, ctx
}
