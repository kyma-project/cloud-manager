package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func deleteAuthTokenSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.authTokenValue == nil {
		return nil, nil
	}

	logger.Info("Deleting authToken secret")

	err := state.awsClient.DeleteAuthTokenSecret(ctx, *state.authTokenValue.Name)
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error deleting authToken secret", ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
