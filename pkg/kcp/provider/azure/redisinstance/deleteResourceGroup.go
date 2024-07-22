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

func deleteResourceGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.azureRedisInstance != nil {
		return nil, nil
	}

	logger.Info("Deleting Azure Redis resourceGroupName")

	resourceGroupName := "cm.redis." + state.ObjAsRedisInstance().Name

	error := state.client.DeleteResourceGroup(ctx, resourceGroupName)
	if error != nil {
		if azuremeta.IsNotFound(error) {
			return nil, nil
		}

		logger.Error(error, "Error deleting Azure resource group")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonCanNotDeleteResourceGroup,
			Message: fmt.Sprintf("Failed deleting AzureRedis resource group: %s", error),
		})
		error = state.UpdateObjStatus(ctx)
		if error != nil {
			return composed.LogErrorAndReturn(error,
				"Error updating RedisInstance status due failed azure redis resource group deleting",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	state.resourceGroup.State = "Deleting"

	return nil, nil
}
