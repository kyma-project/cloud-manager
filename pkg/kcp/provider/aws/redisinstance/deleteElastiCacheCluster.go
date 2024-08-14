package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
)

func deleteElastiCacheCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.elastiCacheCluster == nil {
		return nil, nil
	}
	cacheState := ptr.Deref(state.elastiCacheCluster.Status, "")
	if cacheState == awsmeta.ElastiCache_DELETING {
		return nil, nil
	}

	logger.
		WithValues("elastiCacheCluster", ptr.Deref(state.elastiCacheCluster.ReplicationGroupId, "")).
		Info("Deleting elasti cache cluster")

	err := state.awsClient.DeleteElastiCacheClaster(ctx, ptr.Deref(state.elastiCacheCluster.ReplicationGroupId, ""))
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error deleting elasti cache cluster", ctx)
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
