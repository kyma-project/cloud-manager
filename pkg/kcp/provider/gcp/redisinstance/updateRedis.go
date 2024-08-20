package redisinstance

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	redisInstance := state.ObjAsRedisInstance()

	if state.gcpRedisInstance == nil {
		return composed.StopWithRequeue, nil
	}

	if !state.ShouldUpdateRedisInstance() {
		return nil, nil
	}

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
	err = state.memorystoreClient.UpdateRedisInstance(ctx, state.gcpRedisInstance, state.updateMask)

	if err != nil {
		logger.Error(err, "Error updating GCP Redis")
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonGcpError,
			Message: fmt.Sprintf("Failed updating GcpRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed gcp redis creation",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(5 * util.Timing.T1000ms()), nil
}
