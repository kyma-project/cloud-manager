package redisinstance

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	armRedis "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureUtil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
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
	resourceGroupName := azureUtil.GetResourceGroupName("redis", state.ObjAsRedisInstance().Name)

	redisInstanceName := state.ObjAsRedisInstance().Name
	error := state.client.CreateRedisInstance(
		ctx,
		resourceGroupName,
		redisInstanceName,
		getCreateParams(state),
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

func getCreateParams(state *State) armRedis.CreateParameters {
	createProperties := &armRedis.CreateProperties{
		EnableNonSSLPort: to.Ptr(state.ObjAsRedisInstance().Spec.Instance.Azure.EnableNonSslPort),
		SKU: &armRedis.SKU{
			Name:     to.Ptr(armRedis.SKUNamePremium),
			Capacity: to.Ptr[int32](int32(state.ObjAsRedisInstance().Spec.Instance.Azure.SKU.Capacity)),
			Family:   to.Ptr(armRedis.SKUFamilyP),
		},
		RedisConfiguration: state.ObjAsRedisInstance().Spec.Instance.Azure.RedisConfiguration.GetRedisConfig(),
	}

	if state.ObjAsRedisInstance().Spec.Instance.Azure.ShardCount != 0 {
		createProperties.ShardCount = to.Ptr[int32](int32(state.ObjAsRedisInstance().Spec.Instance.Azure.ShardCount))
	}
	if state.ObjAsRedisInstance().Spec.Instance.Azure.ReplicasPerPrimary != 0 {
		createProperties.ReplicasPerPrimary = to.Ptr[int32](int32(state.ObjAsRedisInstance().Spec.Instance.Azure.ReplicasPerPrimary))
	}
	if state.ObjAsRedisInstance().Spec.Instance.Azure.RedisVersion != "" {
		createProperties.RedisVersion = to.Ptr(state.ObjAsRedisInstance().Spec.Instance.Azure.RedisVersion)
	}

	createParameters := armRedis.CreateParameters{
		Location:   to.Ptr(state.ObjAsRedisInstance().Spec.Instance.Azure.Location),
		Properties: createProperties,
	}
	return createParameters
}
