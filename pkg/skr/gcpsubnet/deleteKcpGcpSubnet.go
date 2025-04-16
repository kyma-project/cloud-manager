package gcpsubnet

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteKcpGcpSubnet(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.KcpGcpSubnet == nil {
		return nil, nil
	}

	if composed.IsMarkedForDeletion(state.KcpGcpSubnet) {
		return nil, nil
	}

	gcpSubnet := state.ObjAsGcpSubnet()

	err, _ := composed.UpdateStatus(gcpSubnet).
		SetCondition(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeDeleting,
			Status:  metav1.ConditionTrue,
			Reason:  cloudresourcesv1beta1.ConditionReasonDeletingInstance,
			Message: fmt.Sprintf("Deleting GcpSubnet %s", state.Name()),
		}).
		ErrorLogMessage("Error setting ConditionReasonDeletingInstance condition on GcpSubnet").
		SuccessErrorNil().
		FailedError(composed.StopWithRequeue).
		Run(ctx, state)
	if err != nil {
		return err, nil
	}

	logger.Info("Deleting KCP GcpSubnet for GcpSubnet")

	err = state.KcpCluster.K8sClient().Delete(ctx, state.KcpGcpSubnet)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting KCP GcpSubnet for GcpSubnet", composed.StopWithRequeue, ctx)
	}

	gcpSubnet.Status.State = cloudresourcesv1beta1.StateDeleting
	err = state.UpdateObjStatus(ctx)

	if err != nil {
		return composed.LogErrorAndReturn(err, "Failed status update on GCP GcpSubnet", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
