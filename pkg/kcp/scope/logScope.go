package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func logScope(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	scopeFound := predicateScopeExists()(ctx, state)
	scopeResourceVersion := state.ObjAsScope().ResourceVersion
	shouldDisable := predicateShouldDisable()(ctx, state)
	shouldEnable := predicateShouldEnable()(ctx, state)
	scopeCreateOrUpdateNeeded := predicateScopeCreateOrUpdateNeeded()(ctx, state)

	logger = logger.
		WithValues(
			"scopeFound", scopeFound,
			"scopeResourceVersion", scopeResourceVersion,
			"shouldDisable", shouldDisable,
			"shouldEnable", shouldEnable,
			"scopeCreateOrUpdateNeeded", scopeCreateOrUpdateNeeded,
		)

	ctx = composed.LoggerIntoCtx(ctx, logger)

	logger.Info("Scope state")

	return nil, ctx
}
