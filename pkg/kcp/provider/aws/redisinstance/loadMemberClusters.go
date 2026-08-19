package redisinstance

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadMemberClusters(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.elastiCacheReplicationGroup == nil {
		return nil, ctx
	}

	var elastiCacheClusters = []elasticachetypes.CacheCluster{}

	for _, memberClusterId := range state.elastiCacheReplicationGroup.MemberClusters {
		clusters, err := state.awsClient.DescribeElastiCacheCluster(ctx, memberClusterId)
		if err != nil {
			logger := logger.WithValues("memberClusterId", memberClusterId)
			return composed.LogErrorAndReturn(err, "failed to describe cluster", composed.StopWithRequeueDelay(5*util.Timing.T10000ms()), composed.LoggerIntoCtx(ctx, logger))
		}

		elastiCacheClusters = append(elastiCacheClusters, clusters...)
	}
	state.memberClusters = elastiCacheClusters

	// Empty member list is not an error; defer readiness/terminal state to waitElastiCacheAvailable.
	if len(elastiCacheClusters) < 1 {
		logger.Info("Replication group has no member clusters yet; deferring to waitElastiCacheAvailable")
	}

	return nil, ctx
}
