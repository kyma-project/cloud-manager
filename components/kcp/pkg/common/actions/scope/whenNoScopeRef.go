package scope

import (
	"context"
	"github.com/kyma-project/cloud-resources-manager/components/lib/composed"
)

func whenNoScopeRef(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	if state.Scope() != nil {
		return nil, nil
	}

	return composed.ComposeActions(
		"whenNoScopeRef",
		findScope,
		composed.BreakIf(func(_ context.Context, _ composed.State) bool {
			return state.Scope() == nil // no scope found, break this flow
		}),
		// scope is found, continue the flow and update the object's scopeRef
		updateScopeRef,
		composed.StopWithRequeueAction,
	)(ctx, st)
}
