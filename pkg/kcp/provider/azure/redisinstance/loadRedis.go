package redisinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureUtil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
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

	if state.resourceGroup == nil {
		logger.Info("Skip the Redis load, resource group is not present and needs to be created (or less possible, loaded)")
		return nil, nil
	}

	logger.Info("Loading Azure Redis")

	redisInstanceName := state.ObjAsRedisInstance().Name
	resourceGroupName := azureUtil.GetResourceGroupName("redis", state.ObjAsRedisInstance().Name)

	redisInstance, err := state.client.GetRedisInstance(ctx, resourceGroupName, redisInstanceName)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, nil
		}

		logger.Error(err, "Error loading Azure Redis")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed loading AzureRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
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
