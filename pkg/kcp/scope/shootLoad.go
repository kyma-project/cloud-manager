package scope

import (
	"context"

	gardenerapicore "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/types"
)

func shootLoad(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	shoot := &gardenerapicore.Shoot{}
	err := state.gardenerClient.Get(ctx, types.NamespacedName{
		Namespace: state.shootNamespace,
		Name:      state.shootName,
	}, shoot)
	if err != nil {
		ctx = composed.LoggerIntoCtx(ctx, logger.WithValues(
			"shootNamespace", state.shootNamespace,
			"shootName", state.shootName,
		))
		return composed.LogErrorAndReturn(err, "Error getting Shoot", composed.StopWithRequeue, ctx)
	}

	state.shoot = shoot

	logger.Info("Shoot loaded")

	return nil, ctx
}
