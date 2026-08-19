package redisinstance

import (
	"context"

	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func modifyPreferredMaintenanceWindow(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisInstance := state.ObjAsRedisInstance()

	if state.elastiCacheReplicationGroup == nil {
		return composed.StopWithRequeue, nil
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

	// No members means no nodes to modify; not an error.
	if len(elastiCacheClusters) < 1 {
		logger.Info("Replication group has no member clusters; skipping preferred maintenance window modification")
		return nil, ctx
	}

	currentPreferredMaintenanceWindow := ptr.Deref(elastiCacheClusters[0].PreferredMaintenanceWindow, "")
	desiredPreferredMaintenanceWindow := ptr.Deref(redisInstance.Spec.Instance.Aws.PreferredMaintenanceWindow, "")

	if currentPreferredMaintenanceWindow == desiredPreferredMaintenanceWindow {
		return nil, ctx
	}
	if desiredPreferredMaintenanceWindow == "" {
		return nil, ctx
	}

	state.UpdatePreferredMaintenanceWindow(desiredPreferredMaintenanceWindow)

	return nil, ctx
}
