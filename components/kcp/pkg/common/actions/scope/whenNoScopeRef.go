package scope

import (
	"context"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

func whenNoScopeRef(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)
	if state.Scope() != nil {
		return nil, nil
	}

	return composed.ComposeActions(
		"whenNoScopeRef",
		findScope,
		// TODO: update scope ref and stop with requeue only if scope if found, otherwise, continue to the next whenNoScopeCreated
		updateScopeRef,
		composed.StopWithRequeueAction,
	)(ctx, st)
}
