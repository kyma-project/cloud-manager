package vpcpeering

import (
	"context"
	"fmt"

	"github.com/alibabacloud-go/tea/tea"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func waitVpcPeeringActive(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	obj := state.ObjAsVpcPeering()

	info, err := state.client.GetVpcPeerConnection(ctx, obj.Status.Id)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error getting AliCloud VpcPeerConnection status", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	if info == nil {
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	state.vpcPeering = info
	status := tea.StringValue(info.Status)

	if status == "Activated" {
		logger.Info("AliCloud VpcPeerConnection is Activated")
		return nil, ctx
	}

	// Deleting/Deleted: still in-flight or already gone; the delete pipeline handles these.
	if status == "Deleting" || status == "Deleted" {
		logger.Info("AliCloud VpcPeerConnection is being deleted", "status", status)
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}

	// Rejected or Expired: unrecoverable — surface an error condition so the operator can act.
	if status == "Rejected" || status == "Expired" {
		logger.Info("AliCloud VpcPeerConnection is in unrecoverable state", "status", status)
		meta.SetStatusCondition(obj.Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
			Message: fmt.Sprintf("VpcPeerConnection is in unrecoverable state: %s", status),
		})
		obj.Status.State = string(cloudcontrolv1beta1.StateError)
		return composed.PatchStatus(obj).
			ErrorLogMessage("Error patching AliCloud VpcPeering status after terminal state").
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	logger.Info("AliCloud VpcPeerConnection waiting", "status", status)
	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
