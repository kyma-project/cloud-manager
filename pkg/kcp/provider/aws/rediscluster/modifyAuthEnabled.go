package rediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func modifyAuthEnabled(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	redisInstance := state.ObjAsRedisCluster()
	logger := composed.LoggerFromCtx(ctx)

	if state.elastiCacheReplicationGroup == nil {
		return composed.StopWithRequeue, nil
	}

	currentAuthEnabled := ptr.Deref(state.elastiCacheReplicationGroup.AuthTokenEnabled, false)
	desiredAuthEnabled := redisInstance.Spec.Instance.Aws.AuthEnabled

	if currentAuthEnabled == desiredAuthEnabled && len(state.elastiCacheReplicationGroup.UserGroupIds) == 0 {
		return nil, ctx
	}

	if desiredAuthEnabled && state.authTokenValue == nil {
		return composed.StopWithRequeue, nil
	}

	if !desiredAuthEnabled && state.authTokenValue != nil {
		logger.Info("Deleting authToken secret")
		err := state.awsClient.DeleteAuthTokenSecret(ctx, *state.authTokenValue.Name)
		if err != nil {
			logger.Error(err, "Error deleting authToken secret")
			meta.SetStatusCondition(redisInstance.Conditions(), metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  cloudcontrolv1beta1.ReasonCloudProviderError,
				Message: "Failed to delete authToken secret",
			})
			redisInstance.Status.State = cloudcontrolv1beta1.StateError
			updateErr := state.UpdateObjStatus(ctx)
			if updateErr != nil {
				return composed.LogErrorAndReturn(updateErr,
					"Error updating RedisCluster status due failed authToken secret deletion",
					composed.StopWithRequeueDelay(util.Timing.T10000ms()),
					ctx,
				)
			}

			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
		}
	}

	state.UpdateAuthEnabled(desiredAuthEnabled)

	return nil, ctx
}
