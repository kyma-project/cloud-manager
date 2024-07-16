package redisinstance

import (
	"context"
	"fmt"
	"github.com/googleapis/gax-go/v2/apierror"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.azureRedisInstance != nil {
		logger.Info("Azure Redis already loaded")
		return nil, nil
	}

	logger.Info("Loading Azure Redis")

	redisInstanceName := state.ObjAsRedisInstance().Name

	redisInstance, error := state.client.GetRedisInstance(ctx, "phoenix-resource-group-1", redisInstanceName)
	if error != nil {
		if apiErr, ok := error.(*apierror.APIError); ok {
			if apiErr.GRPCStatus().Code() == codes.NotFound {
				return nil, nil
			}
		}

		logger.Error(error, "Error loading Azure Redis")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed loading AzureRedis: %s", error),
		})
		error = state.UpdateObjStatus(ctx)
		if error != nil {
			return composed.LogErrorAndReturn(error,
				"Error updating RedisInstance status due failed azure redis loading",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	state.azureRedisInstance = redisInstance

	return nil, nil
}
