package redisinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func upgradeRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	logger := composed.LoggerFromCtx(ctx)
	redisInstance := state.ObjAsRedisInstance()

	if state.gcpRedisInstance == nil {
		return composed.StopWithRequeue, nil
	}

	if !state.ShouldUpgradeRedisInstance() {
		return nil, ctx
	}

	logger.Info("Removing ready state to begin upgrade")
	meta.RemoveStatusCondition(redisInstance.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
	err := state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err,
			"Error upgrading RedisInstance status",
			composed.StopWithRequeueDelay(util.Timing.T10000ms()),
			ctx,
		)
	}

	logger.Info("Updating redis")
	gcpScope := state.Scope().Spec.Scope.Gcp
	region := state.Scope().Spec.Region
	err = state.memorystoreClient.UpgradeRedisInstance(ctx, gcpScope.Project, region, state.GetRemoteRedisName(), redisInstance.Spec.Instance.Gcp.RedisVersion)

	if err != nil {
		logger.Error(err, "Error updating GCP Redis")
		meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
			Message: "Failed to upgrade RedisInstance",
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error upgrading RedisInstance status due failed gcp redis creation",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(30 * util.Timing.T1000ms()), nil
}
