package azurerediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitAuthSecretDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.AuthSecret == nil {
		logger.Info("Auth Secret is deleted")
		return nil, nil
	}

	logger.Info("Waiting for Auth Secret to be deleted")

	return composed.StopWithRequeueDelay(2 * util.Timing.T100ms()), nil
}
