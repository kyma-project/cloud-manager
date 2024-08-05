package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadParameterGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.parameterGroup != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	list, err := state.awsClient.DescribeElastiCacheParameterGroup(ctx, GetAwsElastiCacheParameterGroupName(state.Obj().GetName()))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error listing parameter groups", ctx)
	}

	if len(list) > 0 {
		state.parameterGroup = &list[0]
		logger = logger.WithValues("parameterGroupName", ptr.Deref(state.parameterGroup.CacheParameterGroupName, ""))
		logger.Info("ElastiCache parameter group found and loaded")
		return nil, composed.LoggerIntoCtx(ctx, logger)
	}

	logger.Info("ElastiCache parameter group not found")

	return nil, nil
}
