package redisinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

func removeReadyCondition(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	redisInstance := state.ObjAsRedisInstance()

	readyCond := meta.FindStatusCondition(*redisInstance.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	if readyCond == nil {
		return nil, nil
	}

	logger.Info("Removing Ready condition")

	meta.RemoveStatusCondition(redisInstance.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	redisInstance.Status.State = cloudcontrolv1beta1.DeletingState
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating RedisInstance status after removing Ready condition", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
