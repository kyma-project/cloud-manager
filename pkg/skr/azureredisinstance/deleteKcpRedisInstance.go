package azureredisinstance

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteKcpRedisInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpRedisInstance == nil {
		return nil, nil
	}

	if composed.IsMarkedForDeletion(state.KcpRedisInstance) {
		return nil, nil
	}

	redisInstance := state.ObjAsAzureRedisInstance()

	err, _ := composed.UpdateStatus(redisInstance).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeDeleting,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonDeletingInstance,
			Message: fmt.Sprintf("Deleting RedisInstance %s", state.Name()),
		}).
		ErrorLogMessage("Error setting ConditionReasonDeletingInstance condition on AzureRedisInstance").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
	if err != nil {
		return err, nil
	}

	logger.Info("Deleting KCP RedisInstance for AzureRedisInstance")

	err = state.KcpCluster.K8sClient().Delete(ctx, state.KcpRedisInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP RedisInstance for AzureRedisInstance", composed.StopWithRequeue, ctx)
	}

	redisInstance.Status.State = cloudresourcesv1beta1.StateDeleting
	err = state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Failed status update on GCP RedisInstance", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
