package operations

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// CreateInstance creates a new GCP Filestore instance.
// TODO: Implement this action.
func CreateInstance(ctx context.Context, st composed.State) (error, context.Context) {
	// TODO: Implementation placeholder
	// This will:
	// 1. Build GCP instance spec from CRD
	// 2. Call client.CreateInstance()
	// 3. Store pending operation
	// 4. Set state to creating
	// 5. Return nil, ctx to continue
	return nil, ctx
}
