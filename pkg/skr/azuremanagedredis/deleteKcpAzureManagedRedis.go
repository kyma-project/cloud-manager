package azuremanagedredis

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteKcpAzureManagedRedis(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpAzureManagedRedis == nil {
		return nil, ctx
	}

	if composed.IsMarkedForDeletion(state.KcpAzureManagedRedis) {
		return nil, ctx
	}

	amr := state.ObjAsAzureManagedRedis()

	err, _ := composed.UpdateStatus(amr).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeDeleting,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonDeletingInstance,
			Message: fmt.Sprintf("Deleting AzureManagedRedis %s", state.Name()),
		}).
		ErrorLogMessage("Error setting ConditionReasonDeletingInstance condition on AzureManagedRedis").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
	if err != nil {
		return err, ctx
	}

	logger.Info("Deleting KCP AzureManagedRedis for SKR AzureManagedRedis")

	err = state.KcpCluster.K8sClient().Delete(ctx, state.KcpAzureManagedRedis)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP AzureManagedRedis", composed.StopWithRequeue, ctx)
	}

	amr.Status.State = cloudresourcesv1beta1.StateDeleting
	if err := state.UpdateObjStatus(ctx); err != nil {
		return composed.LogErrorAndReturn(err, "Failed status update on SKR AzureManagedRedis", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
