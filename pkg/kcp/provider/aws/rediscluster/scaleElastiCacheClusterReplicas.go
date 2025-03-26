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

func scaleElastiCacheClusterReplicas(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.elastiCacheReplicationGroup == nil {
		return nil, nil
	}

	if state.IsReplicaCountUpToDate() {
		return nil, nil
	}

	logger.Info("Updating replica count...")

	err := state.awsClient.ModifyElastiCacheClusterReplicaConfiguration(ctx, client.RescaleElastiCacheClusterReplicaOptions{
		ReplicationGroupId:  ptr.Deref(state.elastiCacheReplicationGroup.ReplicationGroupId, ""),
		DesiredReplicaCount: state.ObjAsRedisCluster().Spec.Instance.Aws.ReplicasPerShard,
		ReplicasToRemove:    state.GetReplicasForRemoval(),
	})

	if err != nil {
		logger.Error(err, "Error updating AWS Redis")
		meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to scale RedisCluster replicas",
		})
		state.ObjAsRedisCluster().Status.State = cloudcontrolv1beta1.StateError
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed replica scaling",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T300000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
