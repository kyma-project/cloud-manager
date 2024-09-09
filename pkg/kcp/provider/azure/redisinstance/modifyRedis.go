package redisinstance

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armRedis "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureUtil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func modifyRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	requestedAzureRedisInstance := state.ObjAsRedisInstance()

	if !meta.IsStatusConditionTrue(requestedAzureRedisInstance.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady) {
		return nil, nil
	}

	if state.azureRedisInstance == nil {
		return nil, nil
	}

	updateParams, capacityChanged := getUpdateParams(state)

	if !capacityChanged {
		return nil, nil
	}

	resourceGroupName := azureUtil.GetResourceGroupName("redis", state.ObjAsRedisInstance().Name)
	logger.Info("Detected modified Redis configuration")
	err := state.client.UpdateRedisInstance(
		ctx,
		resourceGroupName,
		requestedAzureRedisInstance.Name,
		updateParams,
	)

	if err != nil {
		logger.Error(err, "Error updating Azure Redis")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed to modify AzureRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisInstance status due failed azure redis update",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}

func getUpdateParams(state *State) (armRedis.UpdateParameters, bool) {

	requestedAzureRedisInstance := state.ObjAsRedisInstance()
	capacityChanged := int(*state.azureRedisInstance.Properties.SKU.Capacity) != requestedAzureRedisInstance.Spec.Instance.Azure.SKU.Capacity
	updateParameters := armRedis.UpdateParameters{}

	if !capacityChanged {
		return updateParameters, false
	}

	updateProperties := &armRedis.UpdateProperties{
		SKU: &armRedis.SKU{
			Name:     to.Ptr(armRedis.SKUNamePremium),
			Capacity: to.Ptr[int32](int32(state.ObjAsRedisInstance().Spec.Instance.Azure.SKU.Capacity)),
			Family:   to.Ptr(armRedis.SKUFamilyP),
		},
	}

	updateParameters.Properties = updateProperties

	return updateParameters, capacityChanged
}
