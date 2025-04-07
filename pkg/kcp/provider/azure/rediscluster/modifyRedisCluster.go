package rediscluster

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func modifyRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	requestedAzureRedisCluster := state.ObjAsRedisCluster()

	if !meta.IsStatusConditionTrue(requestedAzureRedisCluster.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady) {
		return nil, nil
	}

	if state.azureRedisCluster == nil {
		return nil, nil
	}

	updateParams, configChanged := getUpdateParams(state)

	if !configChanged {
		return nil, nil
	}

	resourceGroupName := state.resourceGroupName
	logger.Info("Detected modified Redis configuration")
	err := state.client.UpdateRedisInstance(
		ctx,
		resourceGroupName,
		requestedAzureRedisCluster.Name,
		updateParams,
	)

	if err != nil {
		logger.Error(err, "Error updating Azure Redis")
		meta.SetStatusCondition(state.ObjAsRedisCluster().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: fmt.Sprintf("Failed to modify AzureRedis: %s", err),
		})
		err = state.UpdateObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating RedisCluster status due failed azure redis update",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}

		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
	}

	return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
}

func getUpdateParams(state *State) (armredis.UpdateParameters, bool) {

	requestedAzureRedisCluster := state.ObjAsRedisCluster()
	capacityChanged := int(*state.azureRedisCluster.Properties.SKU.Capacity) != requestedAzureRedisCluster.Spec.Instance.Azure.SKU.Capacity
	shardCountChanged := int(*state.azureRedisCluster.Properties.ShardCount) != requestedAzureRedisCluster.Spec.Instance.Azure.ShardCount
	replicasChanged := int(*state.azureRedisCluster.Properties.ReplicasPerPrimary) != requestedAzureRedisCluster.Spec.Instance.Azure.ReplicasPerPrimary
	redisVersionChanged := *state.azureRedisCluster.Properties.RedisVersion != requestedAzureRedisCluster.Spec.Instance.Azure.RedisVersion
	paramsChanged := capacityChanged || shardCountChanged || replicasChanged || redisVersionChanged

	updateParameters := armredis.UpdateParameters{}
	if !paramsChanged {
		return updateParameters, false
	}

	updateProperties := &armredis.UpdateProperties{
		SKU: &armredis.SKU{
			Capacity: to.Ptr[int32](int32(requestedAzureRedisCluster.Spec.Instance.Azure.SKU.Capacity)),
			Name:     to.Ptr(armredis.SKUNamePremium),
			Family:   to.Ptr(armredis.SKUFamilyP),
		},
		ShardCount:         to.Ptr[int32](int32(requestedAzureRedisCluster.Spec.Instance.Azure.ShardCount)),
		ReplicasPerPrimary: to.Ptr[int32](int32(requestedAzureRedisCluster.Spec.Instance.Azure.ReplicasPerPrimary)),
		RedisVersion:       to.Ptr(requestedAzureRedisCluster.Spec.Instance.Azure.RedisVersion),
	}

	updateParameters.Properties = updateProperties

	return updateParameters, true
}
