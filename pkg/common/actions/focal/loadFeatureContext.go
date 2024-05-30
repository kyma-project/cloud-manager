package focal

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/feature"
)

func loadFeatureContext(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(State)

	ctx = feature.ContextBuilderFromCtx(ctx).
		LoadFromScope(state.Scope()).
		LoadFromKyma(state.Kyma()).
		Build(ctx)

	logger := feature.DecorateLogger(ctx, composed.LoggerFromCtx(ctx))
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
