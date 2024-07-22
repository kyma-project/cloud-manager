package redisinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureUtil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitResourceGroupDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroup == nil {
		return nil, nil
	}

	if state.resourceGroup.State == "Deleting" {
		resourceGroupName := azureUtil.GetResourceGroupName("redis", state.ObjAsRedisInstance().Name)
		resourceGroupExists, error := state.client.ResourceGroupExists(ctx, resourceGroupName)

		if error != nil {
			logger.Error(error, "Error loading Azure resource group while checking if it is deleted")
		}

		if resourceGroupExists == false {
			logger.Info("Deleting Azure Redis resourceGroup done")
			return nil, nil
		}
	}

	logger.Info("Azure Redis resource group is still being deleted, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
