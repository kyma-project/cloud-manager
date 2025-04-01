package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func logScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	gardenerClusterFound := state.gardenerCluster != nil
	gardenerClusterBeingDeleted := composed.IsMarkedForDeletion(state.gardenerCluster)
	scopeFound := composed.IsObjLoaded(ctx, state)
	scopeResourceVersion := state.ObjAsScope().ResourceVersion
	scopeShouldExist := shouldScopeExist(ctx, state)
	scopeCreateOrUpdateNeeded := isScopeCreateOrUpdateNeeded(ctx, state)

	logger = logger.
		WithValues(
			"shootName", state.shootName,
			"gardenerClusterFound", gardenerClusterFound,
			"gardenerClusterBeingDeleted", gardenerClusterBeingDeleted,
			"scopeFound", scopeFound,
			"scopeResourceVersion", scopeResourceVersion,
			"scopeShouldExist", scopeShouldExist,
			"scopeCreateOrUpdateNeeded", scopeCreateOrUpdateNeeded,
		)

	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
