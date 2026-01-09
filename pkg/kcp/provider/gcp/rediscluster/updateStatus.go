package rediscluster

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisCluster := state.ObjAsGcpRedisCluster()

	hasChanged := false

	if len(state.gcpRedisCluster.DiscoveryEndpoints) > 0 {
		discoveryEndpoint := fmt.Sprintf("%s:%d", state.gcpRedisCluster.DiscoveryEndpoints[0].Address, state.gcpRedisCluster.DiscoveryEndpoints[0].Port)
		if redisCluster.Status.DiscoveryEndpoint != discoveryEndpoint {
			redisCluster.Status.DiscoveryEndpoint = discoveryEndpoint
			hasChanged = true
		}
	}

	if redisCluster.Status.CaCert != state.caCerts {
		redisCluster.Status.CaCert = state.caCerts
		hasChanged = true
	}

	nodeType := state.gcpRedisCluster.NodeType.String()
	if redisCluster.Status.NodeType != nodeType {
		redisCluster.Status.NodeType = nodeType
		hasChanged = true
	}

	if state.gcpRedisCluster.ShardCount != nil {
		shardCount := *state.gcpRedisCluster.ShardCount
		if redisCluster.Status.ShardCount != shardCount {
			redisCluster.Status.ShardCount = shardCount
			hasChanged = true
		}
	}

	if state.gcpRedisCluster.ReplicaCount != nil {
		replicasPerShard := *state.gcpRedisCluster.ReplicaCount
		if redisCluster.Status.ReplicasPerShard != replicasPerShard {
			redisCluster.Status.ReplicasPerShard = replicasPerShard
			hasChanged = true
		}
	}

	hasReadyCondition := meta.FindStatusCondition(redisCluster.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
	hasReadyStatusState := redisCluster.Status.State == cloudcontrolv1beta1.StateReady

	if !hasChanged && hasReadyCondition && hasReadyStatusState {
		composed.LoggerFromCtx(ctx).Info("GcpRedisCluster status fields are already up-to-date, StopAndForget-ing")
		return composed.StopAndForget, nil
	}

	redisCluster.Status.State = cloudcontrolv1beta1.StateReady
	return composed.UpdateStatus(redisCluster).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "Redis instance is ready",
		}).
		ErrorLogMessage("Error updating KCP GcpRedisCluster status after setting Ready condition").
		SuccessLogMsg("KCP GcpRedisCluster is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
