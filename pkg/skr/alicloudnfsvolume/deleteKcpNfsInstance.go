package alicloudnfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteKcpNfsInstance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	if state.KcpNfsInstance == nil || composed.IsMarkedForDeletion(state.KcpNfsInstance) {
		// already marked for deletion
		return nil, nil
	}

	err, _ := composed.UpdateStatus(state.ObjAsAlicloudNfsVolume()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeDeleting,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonDeletingInstance,
			Message: fmt.Sprintf("Deleting NfsInstance %s", state.Name()),
		}).
		ErrorLogMessage("Error setting ConditionReasonDeletingInstance condition on AlicloudNfsVolume").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
	if err != nil {
		return err, nil
	}

	logger.Info("Deleting KCP NfsInstance for AlicloudNfsVolume")

	err = state.KcpCluster.K8sClient().Delete(ctx, state.KcpNfsInstance)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP NfsInstance for AlicloudNfsVolume", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
