package rediscluster

import (
	"context"
	"fmt"

	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisCluster := state.ObjAsRedisCluster()
	hasChanged := false

	primaryEndpoint := fmt.Sprintf(
		"%s:%d",
		*state.azureRedisCluster.Properties.HostName,
		*state.azureRedisCluster.Properties.SSLPort,
	)
	if redisCluster.Status.DiscoveryEndpoint != primaryEndpoint {
		redisCluster.Status.DiscoveryEndpoint = primaryEndpoint
		hasChanged = true
	}
	resourceGroupName := state.resourceGroupName
	keys, err := state.client.GetRedisInstanceAccessKeys(ctx, resourceGroupName, state.ObjAsRedisCluster().Name)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error retrieving Azure RedisCluster access keys", composed.StopWithRequeue, ctx)
	}

	authString := ""
	if state.azureRedisCluster != nil {
		authString = pie.First(keys)
	}
	if redisCluster.Status.AuthString != authString {
		redisCluster.Status.AuthString = authString
		hasChanged = true
	}

	if state.azureRedisCluster.Properties != nil && state.azureRedisCluster.Properties.SKU != nil {
		cacheNodeType := fmt.Sprintf("%d", *state.azureRedisCluster.Properties.SKU.Capacity)
		if redisCluster.Status.CacheNodeType != cacheNodeType {
			redisCluster.Status.CacheNodeType = cacheNodeType
			hasChanged = true
		}
	}

	if state.azureRedisCluster.Properties != nil {
		if state.azureRedisCluster.Properties.ShardCount != nil {
			shardCount := *state.azureRedisCluster.Properties.ShardCount
			if redisCluster.Status.ShardCount != shardCount {
				redisCluster.Status.ShardCount = shardCount
				hasChanged = true
			}
		}

		if state.azureRedisCluster.Properties.ReplicasPerPrimary != nil {
			replicasPerShard := *state.azureRedisCluster.Properties.ReplicasPerPrimary
			if redisCluster.Status.ReplicasPerShard != replicasPerShard {
				redisCluster.Status.ReplicasPerShard = replicasPerShard
				hasChanged = true
			}
		} else if state.azureRedisCluster.Properties.ReplicasPerMaster != nil {
			replicasPerShard := *state.azureRedisCluster.Properties.ReplicasPerMaster
			if redisCluster.Status.ReplicasPerShard != replicasPerShard {
				redisCluster.Status.ReplicasPerShard = replicasPerShard
				hasChanged = true
			}
		}
	}

	hasReadyCondition := meta.FindStatusCondition(redisCluster.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasReadyStatusState := redisCluster.Status.State == cloudcontrolv1beta1.StateReady

	if !hasChanged && hasReadyCondition && hasReadyStatusState {
		composed.LoggerFromCtx(ctx).Info("RedisCluster status fields are already up-to-date, StopAndForget-ing")
		return composed.StopAndForget, nil
	}

	redisCluster.Status.State = cloudcontrolv1beta1.StateReady

	return composed.UpdateStatus(redisCluster).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "Redis Cluster is ready",
		}).
		ErrorLogMessage("Error updating KCP RedisCluster status after setting Ready condition").
		SuccessLogMsg("KCP RedisCluster is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
