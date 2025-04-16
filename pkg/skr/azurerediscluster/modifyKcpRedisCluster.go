package azurerediscluster

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyKcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	azureRedisCluster := state.ObjAsAzureRedisCluster()

	if !meta.IsStatusConditionTrue(azureRedisCluster.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady) {
		return nil, nil
	}

	if state.KcpRedisCluster == nil {
		return nil, nil
	}

	redisSKUCapacity, err := RedisTierToSKUCapacityConverter(azureRedisCluster.Spec.RedisTier)

	if err != nil {
		errMsg := "Failed to map redisTier to SKU Capacity"
		logger.Error(err, errMsg, "redisTier", azureRedisCluster.Spec.RedisTier)
		azureRedisCluster.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(azureRedisCluster).
			SetCondition(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: errMsg,
			}).
			RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
			ErrorLogMessage("Error: updating AzureRedisCluster status with not ready condition due to KCP error").
			SuccessLogMsg("Updated and forgot SKR azureRedisCluster status with Error condition").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	capacityChanged := state.KcpRedisCluster.Spec.Instance.Azure.SKU.Capacity != redisSKUCapacity
	shardCountChanged := state.KcpRedisCluster.Spec.Instance.Azure.ShardCount != int(azureRedisCluster.Spec.ShardCount)
	replicasChanged := state.KcpRedisCluster.Spec.Instance.Azure.ReplicasPerPrimary != int(azureRedisCluster.Spec.ReplicasPerPrimary)
	redisVersionChanged := state.KcpRedisCluster.Spec.Instance.Azure.RedisVersion != azureRedisCluster.Spec.RedisVersion

	paramsChanged := capacityChanged || shardCountChanged || replicasChanged || redisVersionChanged

	if !paramsChanged {
		return nil, nil
	}

	state.KcpRedisCluster.Spec.Instance.Azure.SKU.Capacity = redisSKUCapacity
	state.KcpRedisCluster.Spec.Instance.Azure.ShardCount = int(azureRedisCluster.Spec.ShardCount)
	state.KcpRedisCluster.Spec.Instance.Azure.RedisVersion = azureRedisCluster.Spec.RedisVersion

	logger.Info("Detected modified Redis configuration, updating KCP Redis")
	err = state.KcpCluster.K8sClient().Update(ctx, state.KcpRedisCluster)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP AzureRedisCluster", composed.StopWithRequeue, ctx)
	}

	azureRedisCluster.Status.State = cloudresourcesv1beta1.StateUpdating
	return composed.UpdateStatus(azureRedisCluster).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeProcessing,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionTypeProcessing,
			Message: "Processing the resource modification",
		}).
		RemoveConditions(cloudresourcesv1beta1.ConditionTypeError).
		RemoveConditions(cloudresourcesv1beta1.ConditionTypeReady).
		ErrorLogMessage("Error setting Updating state on AzureRedisCluster").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
