package redisinstance

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadResourceGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroup != nil {
		logger.Info("Azure Redis resourceGroupName already loaded")
		return nil, nil
	}

	logger.Info("Loading Azure Redis resourceGroupName")

	resourceGroupName := "cm.redis." + state.ObjAsRedisInstance().Name
	resourceGroupExists, error := state.client.ResourceGroupExists(ctx, resourceGroupName)
	if error != nil {
		if azuremeta.IsNotFound(error) {
			return nil, nil
		}

		logger.Error(error, "Error loading Azure resource group")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonCanNotLoadResourceGroup,
			Message: fmt.Sprintf("Failed loading AzureRedis resource group: %s", error),
		})
		error = state.UpdateObjStatus(ctx)
		if error != nil {
			return composed.LogErrorAndReturn(error,
				"Error updating RedisInstance status due failed azure redis resource group loading",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	if resourceGroupExists {
		state.resourceGroup = &resourceGroup{Name: resourceGroupName}
	}

	return nil, nil
}
