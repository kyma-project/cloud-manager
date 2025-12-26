package focal

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func New() composed.Action {
	return composed.ComposeActions(
		"focal",
		composed.LoadObj,
		loadScopeFromRef,
		loadFeatureContext,
	)
}

// NewWithOptionalScope is handling a special case KCP kyma Network
// when it should be able to delete itself in case the Scope is deleted.
// Restrain yourself from using this function.
func NewWithOptionalScope() composed.Action {
	return composed.ComposeActions(
		"focalOptionalScope",
		func(ctx context.Context, st composed.State) (error, context.Context) {
			state := st.(State)
			state.setScopeOptional()
			return nil, ctx
		},
		New(),
	)
}
