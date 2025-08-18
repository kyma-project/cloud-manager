package redisinstance

import (
	"cloud.google.com/go/redis/apiv1/redispb"
	"context"
	"errors"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitRedisAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.gcpRedisInstance == nil {
		errorMsg := "error: redis instance is not loaded"
		logger.Error(errors.New(errorMsg), errorMsg)
		redisInstance := st.Obj().(*cloudcontrolv1beta1.RedisInstance)
		redisInstance.Status.State = cloudcontrolv1beta1.StateError
		return composed.UpdateStatus(redisInstance).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonUnknown,
				Message: errorMsg,
			}).
			SuccessError(composed.StopAndForget).
			SuccessLogMsg(errorMsg).
			Run(ctx, st)
	}

	if state.gcpRedisInstance.State == redispb.Instance_READY {
		return nil, ctx
	}

	logger.Info("Redis instance is not ready yet, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
