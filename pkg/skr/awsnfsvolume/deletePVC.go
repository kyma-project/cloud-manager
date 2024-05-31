package awsnfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deletePVC(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	if state.PVC == nil {
		return nil, nil
	}

	if !state.PVC.DeletionTimestamp.IsZero() {
		return nil, nil
	}

	err, _ := composed.UpdateStatus(state.ObjAsAwsNfsVolume()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeDeleting,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonDeletingPVC,
			Message: fmt.Sprintf("Deleting PersistentVolumeClaim %s", state.PVC.Name),
		}).
		ErrorLogMessage("Error setting ConditionReasonDeletingPVC condition on AwsNfsVolume").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
	if err != nil {
		return err, nil
	}

	logger.Info("Deleting PVC for AwsNfsVolume")

	err = state.Cluster().K8sClient().Delete(ctx, state.PVC)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting PVC for AwsNfsVolume", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
