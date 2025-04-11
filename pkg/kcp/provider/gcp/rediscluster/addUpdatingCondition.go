package rediscluster

import (
	"context"

	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func addUpdatingCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisCluster := state.ObjAsGcpRedisCluster()

	if state.gcpRedisCluster == nil {
		return nil, nil
	}

	isModifying := state.gcpRedisCluster.State == clusterpb.Cluster_UPDATING
	hasUpdatingCondition := meta.FindStatusCondition(redisCluster.Status.Conditions, cloudcontrolv1beta1.ConditionTypeUpdating) != nil

	if !isModifying && !hasUpdatingCondition {
		return nil, nil
	}

	if isModifying && hasUpdatingCondition {
		return nil, nil
	}

	if isModifying && !hasUpdatingCondition {
		logger.Info("Adding updating condition to redis cluster.")
		return composed.UpdateStatus(redisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeUpdating,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeUpdating,
				Message: "ElastiCache is updating.",
			}).
			SuccessErrorNil().
			ErrorLogMessage("Failed to add updating condition to redis cluster").
			Run(ctx, st)
	}

	if !isModifying && hasUpdatingCondition {
		logger.Info("Removing updating condition from redis cluster")
		return composed.UpdateStatus(redisCluster).
			RemoveConditions(cloudcontrolv1beta1.ConditionTypeUpdating).
			SuccessErrorNil().
			ErrorLogMessage("Failed to remove updating condition from redis cluster").
			Run(ctx, st)
	}

	return nil, nil
}
