package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func restartElastiCacheCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.elastiCacheCluster == nil {
		return nil, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	if state.elastiCacheCluster.PendingModifiedValues == nil {
		return nil, nil
	}

	for _, memberClusterId := range state.elastiCacheCluster.MemberClusters {
		err := state.awsClient.RestartElastiCacheCluster(ctx, memberClusterId)
		if err != nil {
			logger := logger.WithValues("memberClusterId", memberClusterId)
			return composed.LogErrorAndReturn(err, "failed to restart cluster", composed.StopWithRequeueDelay(5*util.Timing.T10000ms()), composed.LoggerIntoCtx(ctx, logger))
		}
	}

	return composed.StopWithRequeueDelay(10 * util.Timing.T10000ms()), nil
}
