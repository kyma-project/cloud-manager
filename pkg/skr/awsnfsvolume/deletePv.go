package awsnfsvolume

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deletePv(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	if state.Volume == nil {
		return nil, nil
	}

	if !state.Volume.DeletionTimestamp.IsZero() {
		return nil, nil
	}

	err, _ := composed.UpdateStatus(state.ObjAsAwsNfsVolume()).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeDeleting,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonDeletingPV,
			Message: fmt.Sprintf("Deleting PersistentVolume %s", state.Volume.Name),
		}).
		ErrorLogMessage("Error setting ConditionReasonDeletingPV condition on AwsNfsVolume").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
	if err != nil {
		return err, nil
	}

	logger.Info("Deleting PV for AwsNfsVolume")

	err = state.Cluster().K8sClient().Delete(ctx, state.Volume)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting PV for AwsNfsVolume", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
