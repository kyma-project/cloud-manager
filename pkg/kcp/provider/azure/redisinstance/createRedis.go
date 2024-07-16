package redisinstance

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armRedis "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.azureRedisInstance != nil {
		return nil, nil
	}

	logger.Info("Creating Azure Redis")

	redisInstanceName := state.ObjAsRedisInstance().Name
	_, error := state.client.CreateRedisInstance(
		ctx,
		"phoenix-resource-group-1",
		redisInstanceName,
		armRedis.CreateParameters{
			Location: to.Ptr("East US"),
			Properties: &armRedis.CreateProperties{
				RedisVersion: to.Ptr(state.ObjAsRedisInstance().Spec.Instance.Azure.RedisVersion),
				SKU: &armRedis.SKU{
					Name:     to.Ptr(armRedis.SKUNamePremium),
					Capacity: to.Ptr[int32](int32(state.ObjAsRedisInstance().Spec.Instance.Azure.SKU.Capacity)),
					Family:   to.Ptr(armRedis.SKUFamilyP),
				},
			},
		},
	)

	if error != nil {
		logger.Error(error, "Error creating Azure Redis")
		meta.SetStatusCondition(state.ObjAsRedisInstance().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed creating AzureRedis: %s", error),
		})
		error = state.UpdateObjStatus(ctx)
		if error != nil {
			return composed.LogErrorAndReturn(error,
				"Error updating RedisInstance status due failed azure redis creation",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}

	return composed.StopWithRequeue, nil
}
