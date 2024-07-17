package redisinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitRedisDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.azureRedisInstance == nil {
		return nil, nil
	}

	if *state.azureRedisInstance.Properties.ProvisioningState != "Deleting" {
		errorMsg := "Error: unexpected azure redis state"
		redisInstance := st.Obj().(*v1beta1.RedisInstance)
		return composed.UpdateStatus(redisInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ConditionTypeError,
				Message: errorMsg,
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg(errorMsg).
			Run(ctx, st)
	}

	logger.Info("Azure Redis instance is still being deleted, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
