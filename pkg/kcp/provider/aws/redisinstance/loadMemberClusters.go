package redisinstance

import (
	"context"
	"errors"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func loadMemberClusters(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.elastiCacheReplicationGroup == nil {
		return nil, nil
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

	if len(elastiCacheClusters) < 1 {
		return composed.LogErrorAndReturn(errors.New("no replication group clusters found"), "no replication group clusters found", composed.StopWithRequeueDelay(5*util.Timing.T10000ms()), ctx)
	}

	return nil, nil
}
