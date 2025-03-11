package rediscluster

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.azureRedisCluster != nil {
		return nil, nil
	}

	logger.Info("Creating Azure Redis")
	resourceGroupName := state.resourceGroupName

	redisClusterName := state.ObjAsRedisCluster().Name
	err := state.client.CreateRedisInstance(
		ctx,
		resourceGroupName,
		redisClusterName,
		getCreateParams(state),
	)

	if err != nil {
		logger.Error(err, "Error creating Azure Redis")
		meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ReasonFailedCreatingFileSystem,
			Message: fmt.Sprintf("Failed creating AzureRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed azure redis creation",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}

func getCreateParams(state *State) armredis.CreateParameters {
	createProperties := &armredis.CreateProperties{
		SKU: &armredis.SKU{
			Name:     to.Ptr(armredis.SKUNamePremium),
			Capacity: to.Ptr[int32](int32(state.ObjAsRedisCluster().Spec.Instance.Azure.SKU.Capacity)),
			Family:   to.Ptr(armredis.SKUFamilyP),
		},
		RedisConfiguration: state.ObjAsRedisCluster().Spec.Instance.Azure.RedisConfiguration.GetRedisConfig(),
		EnableNonSSLPort:   to.Ptr(false),
	}

	if state.ObjAsRedisCluster().Spec.Instance.Azure.ShardCount != 0 {
		createProperties.ShardCount = to.Ptr[int32](int32(state.ObjAsRedisCluster().Spec.Instance.Azure.ShardCount))
		createProperties.ReplicasPerPrimary = to.Ptr[int32](int32(state.ObjAsRedisCluster().Spec.Instance.Azure.ReplicasPerPrimary))
	}
	if state.ObjAsRedisCluster().Spec.Instance.Azure.RedisVersion != "" {
		createProperties.RedisVersion = to.Ptr(state.ObjAsRedisCluster().Spec.Instance.Azure.RedisVersion)
	}

	createParameters := armredis.CreateParameters{
		Location:   to.Ptr(state.Scope().Spec.Region),
		Properties: createProperties,
	}
	return createParameters
}
