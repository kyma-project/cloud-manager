package rediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	redisCluster := state.ObjAsRedisCluster()

	if state.gcpRedisCluster == nil {
		return composed.StopWithRequeue, nil
	}

	if !state.ShouldUpdateRedisCluster() {
		return nil, nil
	}

	logger.Info("Removing ready state to begin update")
	meta.RemoveStatusCondition(redisCluster.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error updating RedisCluster status",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}

	logger.Info("Updating redis")
	err = state.memorystoreClient.UpdateRedisCluster(ctx, state.gcpRedisCluster, state.updateMask)

	if err != nil {
		logger.Error(err, "Error updating GCP Redis")
		meta.SetStatusCondition(redisCluster.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to update RedisCluster",
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed gcp redis creation",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(30 * util.Timing.T1000ms()), nil
}
