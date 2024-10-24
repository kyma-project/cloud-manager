package redisinstance

import (
	"context"
	"fmt"

	"cloud.google.com/go/redis/apiv1/redispb"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
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
		logger.Error(err, "Error deleting GCP Redis")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonGcpError,
			Message: fmt.Sprintf("Failed deleting GcpRedis: %s", err),
		})
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
