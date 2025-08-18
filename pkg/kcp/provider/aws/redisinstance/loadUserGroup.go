package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
)

func loadUserGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.userGroup != nil {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)

	userGroup, err := state.awsClient.DescribeUserGroup(ctx, GetAwsElastiCacheUserGroupName(state.Obj().GetName()))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error getting user group", ctx)
	}

	if userGroup == nil {
		logger.Info("ElastiCache User group not found")
		return nil, ctx
	}

	state.userGroup = userGroup
	logger.Info("ElastiCache user group found and loaded")

	return nil, ctx
}
