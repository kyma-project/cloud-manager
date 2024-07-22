package redisinstance

import (
	"context"
	"fmt"
	azureUtil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"

	"github.com/googleapis/gax-go/v2/apierror"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/grpc/codes"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.azureRedisInstance == nil {
		return nil, nil
	}

	if *state.azureRedisInstance.Properties.ProvisioningState == "Deleting" {
		return nil, nil // delete is waited in next action
	}

	logger.Info("Deleting Azure Redis")

	redisInstanceName := state.ObjAsRedisInstance().Name
	resourceGroupName := azureUtil.GetResourceGroupName("redis", state.ObjAsRedisInstance().Name)

	err := state.client.DeleteRedisInstance(ctx, resourceGroupName, redisInstanceName)
	if err != nil {
		if apiErr, ok := err.(*apierror.APIError); ok {
			if apiErr.GRPCStatus().Code() == codes.NotFound {
				return nil, nil
			}
		}

		logger.Error(err, "Error deleting Azure Redis")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed deleting AzureRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed azure redis deleting",
				composed.StopWithRequeueDelay((util.Timing.T10000ms())),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
}
