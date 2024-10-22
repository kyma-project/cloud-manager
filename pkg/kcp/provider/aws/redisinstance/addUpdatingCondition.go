package redisinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func addUpdatingCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisInstance := state.ObjAsRedisInstance()

	if state.elastiCacheReplicationGroup == nil {
		return nil, nil
	}

	cacheState := ptr.Deref(state.elastiCacheReplicationGroup.Status, "")
	isModifying := cacheState == awsmeta.ElastiCache_MODIFYING
	hasUpdatingCondition := meta.FindStatusCondition(redisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeUpdating) != nil

	if !isModifying && !hasUpdatingCondition {
		return nil, nil
	}

	if isModifying && hasUpdatingCondition {
		return nil, nil
	}

	if isModifying && !hasUpdatingCondition {
		logger.Info("Adding updating condition to redis instance.")
		return composed.UpdateStatus(redisInstance).
			SetCondition(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeUpdating,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeUpdating,
				Message: "ElastiCache is updating.",
			}).
			SuccessErrorNil().
			ErrorLogMessage("Failed to add updating condition to redis instance").
			Run(ctx, st)
	}

	if !isModifying && hasUpdatingCondition {
		logger.Info("Removing updating condition from redis instance")
		return composed.UpdateStatus(redisInstance).
			RemoveConditions(cloudcontrolv1beta1.ConditionTypeUpdating).
			SuccessErrorNil().
			ErrorLogMessage("Failed to remove updating condition from redis instance").
			Run(ctx, st)
	}

	return nil, nil
}
