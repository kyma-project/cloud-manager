package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// modifyShardCount grows or shrinks the cluster to match the desired ShardCount.
// AddShardingNode and DeleteShardingNode accept a delta (number of shards to
// add or remove), not an absolute target — the delta is computed here.
// ShardCount changes are always issued in a separate pipeline step from
// InstanceClass changes so that the intermediate waitRedisAvailable can
// confirm the instance has stabilised between the two operations.
func modifyShardCount(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	if state.instance == nil {
		return nil, ctx
	}
	kcp := state.ObjAsRedisCluster()
	if kcp.Spec.Instance.Alicloud == nil {
		return nil, ctx
	}
	desired := kcp.Spec.Instance.Alicloud.ShardCount
	current := state.instance.ShardCount

	// current == 0 has two meanings: the API hasn't surfaced the count yet (instance
	// still Creating), or the class is proxy-based (redis.logic.sharding.*) where
	// ShardCount is encoded in the class name and never returned by DescribeInstance.
	// In both cases skip — proxy shard changes are handled by modifyInstanceClass.
	if current == 0 || desired == current {
		return nil, ctx
	}

	var err error
	if desired > current {
		err = state.client.AddShardingNode(ctx, state.instance.InstanceId, desired-current)
	} else {
		err = state.client.DeleteShardingNode(ctx, state.instance.InstanceId, current-desired)
	}
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error modifying AliCloud r-kvstore cluster shard count",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx)
	}

	state.instance = nil
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), ctx
}
