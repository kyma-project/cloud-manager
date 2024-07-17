package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadElastiCacheCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.elastiCacheCluster != nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	list, err := state.awsClient.DescribeElastiCacheCluster(ctx, state.Obj().GetName())
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error listing elasticache clusters", ctx)
	}

	if len(list) > 0 {
		state.elastiCacheCluster = &list[0]
		logger = logger.WithValues("elastiCacheClusterId", ptr.Deref(state.elastiCacheCluster.CacheClusterId, ""))
		logger.Info("ElastiCache cluster found and loaded")
		return nil, composed.LoggerIntoCtx(ctx, logger)
	}

	logger.Info("ElastiCache cluster not found")

	return nil, nil
}
