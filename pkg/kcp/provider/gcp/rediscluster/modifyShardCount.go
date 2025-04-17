package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/ptr"
)

func modifyShardCount(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisCluster := state.ObjAsGcpRedisCluster()

	if state.gcpRedisCluster == nil {
		return composed.StopWithRequeue, nil
	}

	currentShardCount := ptr.Deref(state.gcpRedisCluster.ShardCount, redisCluster.Spec.ShardCount)
	desiredShardCount := redisCluster.Spec.ShardCount

	if currentShardCount == desiredShardCount {
		return nil, nil
	}

	state.UpdateShardCount(desiredShardCount)

	return nil, nil
}
