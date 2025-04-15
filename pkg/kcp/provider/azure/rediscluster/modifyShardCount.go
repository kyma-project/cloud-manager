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

func modifyShardCount(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	requestedAzureRedisCluster := state.ObjAsRedisCluster()

	if !meta.IsStatusConditionTrue(requestedAzureRedisCluster.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady) {
		return nil, nil
	}

	if state.azureRedisCluster == nil {
		return nil, nil
	}

	shardCountChanged := int(*state.azureRedisCluster.Properties.ShardCount) != requestedAzureRedisCluster.Spec.Instance.Azure.ShardCount

	if !shardCountChanged {
		return nil, nil
	}

	resourceGroupName := state.resourceGroupName
	logger.Info("Detected modified Redis configuration - shardCount")
	err := state.client.UpdateRedisInstance(
		ctx,
		resourceGroupName,
		requestedAzureRedisCluster.Name,
		armredis.UpdateParameters{
			Properties: &armredis.UpdateProperties{
				ShardCount: to.Ptr[int32](int32(requestedAzureRedisCluster.Spec.Instance.Azure.ShardCount)),
			},
		},
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
