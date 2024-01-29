package focal

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
)

// loadObj loads object with state.Name into the state.Obj()
//
// Deprecated: Use composed.LoadObj() instead
func loadObj(ctx context.Context, state composed.State) (error, context.Context) {
	return composed.LoadObj(ctx, state)
}
