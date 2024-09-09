package redisinstance

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
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

	if state.resourceGroup == nil {
		return nil, nil
	}

	if *state.resourceGroup.Properties.ProvisioningState == string(armredis.ProvisioningStateDeleting) {
		return nil, nil
	}

	logger.Info("Deleting Azure Redis resourceGroup")

	resourceGroupName := state.resourceGroupName

	err := state.client.DeleteResourceGroup(ctx, resourceGroupName)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, nil
		}

		logger.Error(err, "Error deleting Azure resource group")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonCanNotDeleteResourceGroup,
			Message: fmt.Sprintf("Failed deleting AzureRedis resource group: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed azure redis resource group deleting",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return nil, nil
}
