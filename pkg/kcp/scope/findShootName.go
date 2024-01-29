package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
)

func findShootName(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	if state.kyma == nil {
		logger.Info("No Kyma CR loaded, should not have been called!")
		return composed.StopAndForget, nil
	}

	state.shootName = state.kyma.GetLabels()["kyma-project.io/shoot-name"]

	logger = logger.WithValues("shootName", state.shootName)
	logger.Info("Shoot name found")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
