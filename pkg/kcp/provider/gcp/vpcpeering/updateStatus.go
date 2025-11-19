package vpcpeering

import (
	"context"
	"strings"

	pb "cloud.google.com/go/compute/apiv1/computepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		logger.Info("GCP VPC Peering is marked for deletion")
		return nil, nil
	}

	var op *pb.Operation
	if state.remoteOperation.GetError() != nil {
		op = state.remoteOperation
	}
	if state.localOperation.GetError() != nil {
		op = state.localOperation
	}

	if op != nil && op.GetError() != nil {

		state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected
		if strings.Contains(state.remoteOperation.GetError().String(), "QUOTA_EXCEEDED") {
			return composed.UpdateStatus(state.ObjAsVpcPeering()).SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeQuotaExceeded,
				Status:  "True",
				Reason:  v1beta1.ConditionTypeQuotaExceeded,
				Message: "Error creating Vpc Peering due to quota limits " + op.GetDescription() + ", please check if your vpc quota limits are not exceeded.",
			}).
				ErrorLogMessage("Failed to update status to set quota exceeded on vpc peering").
				FailedError(composed.StopWithRequeue).
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}

		return composed.UpdateStatus(state.ObjAsVpcPeering()).SetExclusiveConditions(metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionTypeError,
			Message: "The cloud provider had an error while creating Vpc Peering" + op.GetDescription(),
		}).
			ErrorLogMessage("The cloud provider had an error while creating Remote Vpc Peering"+op.GetDescription()).
			FailedError(composed.StopWithRequeue).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if meta.IsStatusConditionTrue(
		ptr.Deref(state.ObjAsVpcPeering().Conditions(), []metav1.Condition{}),
		cloudcontrolv1beta1.ConditionTypeReady,
	) {
		return nil, nil
	}

	state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected

	return composed.UpdateStatus(state.ObjAsVpcPeering()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeReady,
			Status:  "True",
			Reason:  cloudcontrolv1beta1.ReasonReady,
			Message: "VpcPeering :" + state.remotePeeringName + " is provisioned",
		}).
		ErrorLogMessage("Error updating VpcPeering success status after setting Ready condition").
		SuccessLogMsg("KPC VpcPeering is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
