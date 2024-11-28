package vpcpeering

import (
	"context"
	ec2Types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func waitPendingAcceptance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	code := state.vpcPeering.Status.Code

	if code == ec2Types.VpcPeeringConnectionStateReasonCodeActive ||
		code == ec2Types.VpcPeeringConnectionStateReasonCodeProvisioning ||
		code == ec2Types.VpcPeeringConnectionStateReasonCodePendingAcceptance {
		return nil, nil
	}

	changed := false

	if state.ObjAsVpcPeering().Status.State != string(code) {
		state.ObjAsVpcPeering().Status.State = string(code)
		changed = true
	}

	if code == ec2Types.VpcPeeringConnectionStateReasonCodeFailed ||
		code == ec2Types.VpcPeeringConnectionStateReasonCodeExpired ||
		code == ec2Types.VpcPeeringConnectionStateReasonCodeRejected ||
		code == ec2Types.VpcPeeringConnectionStateReasonCodeDeleted ||
		code == ec2Types.VpcPeeringConnectionStateReasonCodeDeleting {

		condition := metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedAcceptingVpcPeeringConnection,
			Message: ptr.Deref(state.vpcPeering.Status.Message, ""),
		}

		if composed.AnyConditionChanged(state.ObjAsVpcPeering(), condition) ||
			changed {
			return composed.PatchStatus(state.ObjAsVpcPeering()).
				SetExclusiveConditions(condition).
				ErrorLogMessage("Error updating VpcPeering status while waiting for AWS VPC peering pending-acceptance").
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}

		return composed.StopAndForget, nil
	}

	// code initiating-request
	if changed {
		return composed.PatchStatus(state.ObjAsVpcPeering()).
			ErrorLogMessage("Error setting KCP VpcPeering status state").
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return composed.StopWithRequeue, nil
}
