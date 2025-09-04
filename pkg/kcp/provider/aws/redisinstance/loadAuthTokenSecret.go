package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func loadAuthTokenSecret(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.authTokenValue != nil {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)

	authTokenValue, err := state.awsClient.GetAuthTokenSecretValue(ctx, GetAwsAuthTokenSecretName(state.Obj().GetName()))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error getting auth token", ctx)
	}

	if authTokenValue == nil {
		logger.Info("ElastiCache auth token valuenot found")
		return nil, ctx
	}

	state.authTokenValue = authTokenValue
	logger.Info("ElastiCache auth token value found and loaded")

	return nil, ctx
}
