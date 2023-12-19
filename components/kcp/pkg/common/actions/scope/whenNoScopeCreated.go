package scope

import (
	"context"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

func whenNoScopeCreated(ctx context.Context, state composed.State) (error, context.Context) {
	if state.(State).Scope() != nil {
		return nil, nil
	}
	return composed.ComposeActions(
		"whenNoScopeCreated",
		loadKyma,
		createGardenerClient,
		loadShoot,
		loadGardenerCredentials,
		createScope,
		ensureScopeCommonFields,
		saveScope,
		updateScopeRef,
		// scope is created, requeue now
		composed.StopWithRequeueAction,
	)(ctx, state)
}
