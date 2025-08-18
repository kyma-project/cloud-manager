package redisinstance

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

func modifyElastiCacheReplicationGroup(updateParamsFn func(*State) client.ModifyElastiCacheClusterOptions) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)
		redisInstance := state.ObjAsRedisInstance()

		if state.elastiCacheReplicationGroup == nil {
			return nil, ctx
		}

		logger := composed.LoggerFromCtx(ctx)

		logger.Info("Removing ready state to begin update")
		meta.RemoveStatusCondition(redisInstance.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
		err := state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		logger.Info("Updating redis")
		_, err = state.awsClient.ModifyElastiCacheReplicationGroup(ctx, ptr.Deref(state.elastiCacheReplicationGroup.ReplicationGroupId, ""), updateParamsFn(state))

		if err != nil {
			logger.Error(err, "Error updating AWS Redis")
			meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: "Failed to update RedisInstance",
			})
			redisInstance.Status.State = cloudcontrolv1beta1.StateError
			err = state.UpdateObjStatus(ctx)
			if err != nil {
				return composed.LogErrorAndReturn(err,
					"Error updating RedisInstance status due failed aws redis creation",
					composed.StopWithRequeueDelay((util.Timing.T10000ms())),
					ctx,
				)
			}

			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
		}

		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	}
}

func updateElastiCacheCluster() composed.Action {
	return modifyElastiCacheReplicationGroup(func(s *State) client.ModifyElastiCacheClusterOptions {
		return s.GetModifyElastiCacheClusterOptions()
	})
}

func upgradeElastiCacheCluster() composed.Action {
	return modifyElastiCacheReplicationGroup(func(s *State) client.ModifyElastiCacheClusterOptions {
		return client.ModifyElastiCacheClusterOptions{
			EngineVersion:      ptr.To(s.ObjAsRedisInstance().Spec.Instance.Aws.EngineVersion),
			ParameterGroupName: ptr.To(s.GetUpgradeParamGroupName()),
		}
	})
}

func switchToMainParamGroup() composed.Action {
	return modifyElastiCacheReplicationGroup(func(s *State) client.ModifyElastiCacheClusterOptions {
		return client.ModifyElastiCacheClusterOptions{
			ParameterGroupName: s.parameterGroup.CacheParameterGroupName,
		}
	})
}
