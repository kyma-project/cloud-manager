package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadShoot(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	shoot, err := state.gardenerClient.Shoots(state.shootNamespace).Get(ctx, state.shootName, metav1.GetOptions{})
	if err != nil {
		ctx = composed.LoggerIntoCtx(ctx, logger.WithValues(
			"shootNamespace", state.shootNamespace,
			"shootName", state.shootName,
		))
		return composed.LogErrorAndReturn(err, "Error getting Shoot", composed.StopWithRequeue, ctx)
	}

	state.shoot = shoot

	logger.Info("Shoot loaded")

	return nil, nil
}
