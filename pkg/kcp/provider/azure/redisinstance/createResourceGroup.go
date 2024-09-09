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

func createResourceGroup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroup != nil {
		return nil, nil
	}

	logger.Info("Creating Azure Redis resourceGroup")

	resourceGroupName := azureUtil.GetResourceGroupName("redis", state.ObjAsRedisInstance().Name)
	location := state.Scope().Spec.Region

	err := state.client.CreateResourceGroup(ctx, resourceGroupName, location)
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, nil
		}

		logger.Error(err, "Error crating Azure resource group")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonCanNotCreateResourceGroup,
			Message: fmt.Sprintf("Failed creating AzureRedis resource group: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed azure redis resource group create",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	// we have just created the group, requeue so the Redis can be loaded
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
