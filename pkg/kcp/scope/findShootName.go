package scope

import (
	"context"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func findShootName(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	if state.kyma == nil {
		logger.Info("No Kyma CR loaded, should not have been called!")
		return composed.StopAndForget, nil
	}

	state.shootName = state.kyma.GetLabels()[cloudcontrol1beta1.LabelScopeShootName]

	logger = logger.WithValues("shootName", state.shootName)
	logger.Info("Shoot name found")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
