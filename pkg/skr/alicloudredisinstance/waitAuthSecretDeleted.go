package alicloudredisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// waitAuthSecretDeleted re-checks state.AuthSecret which is reloaded from the
// cluster on every reconcile cycle by loadAuthSecret earlier in the pipeline.
// "Waiting" here means the reconcile terminates and re-runs; it does not poll
// in-process. The secret is considered gone when loadAuthSecret sets
// state.AuthSecret = nil (i.e. Get returns NotFound).
func waitAuthSecretDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.AuthSecret == nil {
		logger.Info("Auth Secret is deleted")
		return nil, ctx
	}

	logger.Info("Waiting for Auth Secret to be deleted")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
