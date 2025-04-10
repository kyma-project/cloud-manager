package redisinstance

import (
	"context"

	"cloud.google.com/go/redis/apiv1/redispb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.gcpRedisInstance == nil {
		return nil, nil
	}

	if state.gcpRedisInstance.State == redispb.Instance_DELETING {
		return nil, nil // delete is waited in next action
	}

	logger.Info("Deleting GCP Redis")

	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region

	err := state.memorystoreClient.DeleteRedisInstance(ctx, gcpScope.Project, region, state.GetRemoteRedisName())
	if err != nil {
		if gcpmeta.IsNotFound(err) {
			logger.Info("target redis instance for delete not found, continuing to next loop")
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
		}

		logger.Error(err, "Error deleting GCP Redis")
		redisInstance := state.ObjAsRedisInstance()
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to delete RedisInstance",
		})
		redisInstance.Status.State = cloudcontrolv1beta1.StateError

		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed gcp redis deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
