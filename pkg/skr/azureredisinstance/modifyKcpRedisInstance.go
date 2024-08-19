package azureredisinstance

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func modifyKcpRedisInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	azureRedisInstance := state.ObjAsAzureRedisInstance()

	if !meta.IsStatusConditionTrue(azureRedisInstance.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady) {
		return nil, nil
	}

	if state.KcpRedisInstance == nil {
		return nil, nil
	}

	capacityChanged := state.KcpRedisInstance.Spec.Instance.Azure.SKU.Capacity != azureRedisInstance.Spec.SKU.Capacity

	if !capacityChanged {
		return nil, nil
	}

	state.KcpRedisInstance.Spec.Instance.Azure.SKU.Capacity = azureRedisInstance.Spec.SKU.Capacity
	logger.Info("Detected modified Redis SKU capacity")
	err := state.KcpCluster.K8sClient().Update(ctx, state.KcpRedisInstance)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP AzureRedisInstance", composed.StopWithRequeue, ctx)
	}

	azureRedisInstance.Status.State = cloudresourcesv1beta1.StateUpdating
	return composed.UpdateStatus(azureRedisInstance).
		ErrorLogMessage("Error setting Updating state on AzureRedisInstance").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
}
