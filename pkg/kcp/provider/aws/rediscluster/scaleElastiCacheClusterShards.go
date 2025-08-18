package rediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func scaleElastiCacheClusterShards(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.elastiCacheReplicationGroup == nil {
		return nil, ctx
	}

	if state.IsShardCountUpToDate() {
		return nil, ctx
	}

	logger.Info("Updating shard count...")

	err := state.awsClient.ModifyElastiCacheClusterShardConfiguration(ctx, client.RescaleElastiCacheClusterShardOptions{
		ReplicationGroupId: ptr.Deref(state.elastiCacheReplicationGroup.ReplicationGroupId, ""),
		DesiredShardCount:  state.ObjAsRedisCluster().Spec.Instance.Aws.ShardCount,
		NodeGroupsToRemove: state.GetShardsForRemoval(),
	})

	if err != nil {
		logger.Error(err, "Error updating AWS Redis")
		meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to scale RedisCluster shards",
		})
		state.ObjAsRedisCluster().Status.State = cloudcontrolv1beta1.StateError
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed shard scaling",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T300000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
