package azurerediscluster

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteKcpRedisCluster(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpRedisCluster == nil {
		return nil, ctx
	}

	if composed.IsMarkedForDeletion(state.KcpRedisCluster) {
		return nil, ctx
	}

	redisCluster := state.ObjAsAzureRedisCluster()

	err, _ := composed.UpdateStatus(redisCluster).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeDeleting,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonDeletingInstance,
			Message: fmt.Sprintf("Deleting RedisCluster %s", state.Name()),
		}).
		ErrorLogMessage("Error setting ConditionReasonDeletingCluster condition on AzureRedisCluster").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
	if err != nil {
		return err, ctx
	}

	logger.Info("Deleting KCP RedisCluster for AzureRedisCluster")

	err = state.KcpCluster.K8sClient().Delete(ctx, state.KcpRedisCluster)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP RedisCluster for AzureRedisCluster", composed.StopWithRequeue, ctx)
	}

	redisCluster.Status.State = cloudresourcesv1beta1.StateDeleting
	err = state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Failed status update on GCP RedisCluster", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
