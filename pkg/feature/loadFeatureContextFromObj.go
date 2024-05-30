package feature

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func LoadFeatureContextFromObj(obj client.Object) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		ctx = ContextBuilderFromCtx(ctx).
			FeatureFromObject(obj, state.Cluster().Scheme()).
			Build(ctx)
		logger := composed.LoggerFromCtx(ctx)
		logger = DecorateLogger(ctx, logger)
		ctx = composed.LoggerIntoCtx(ctx, logger)

		return nil, ctx
	}
}
