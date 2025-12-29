package operations

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// LoadInstance loads the GCP Filestore instance from the API.
// TODO: Implement this action.
func LoadInstance(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Implementation placeholder
	// This will:
	// 1. Get project, location, instance name from state
	// 2. Call client.GetInstance()
	// 3. Handle not found (mark as needs creation)
	// 4. Cache instance in state
	// 5. Return nil, ctx to continue
	return nil, ctx
}
