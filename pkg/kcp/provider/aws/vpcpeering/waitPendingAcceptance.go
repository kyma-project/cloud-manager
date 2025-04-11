package vpcpeering

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func waitPendingAcceptance(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	code := state.vpcPeering.Status.Code

	if code == ec2types.VpcPeeringConnectionStateReasonCodeInitiatingRequest {
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	}

	// can't continue if VPC peering connection is in one of these statuses
	if code == ec2types.VpcPeeringConnectionStateReasonCodeFailed ||
		code == ec2types.VpcPeeringConnectionStateReasonCodeExpired ||
		code == ec2types.VpcPeeringConnectionStateReasonCodeRejected ||
		code == ec2types.VpcPeeringConnectionStateReasonCodeDeleted ||
		code == ec2types.VpcPeeringConnectionStateReasonCodeDeleting {

		changed := false

		if state.ObjAsVpcPeering().Status.State != string(code) {
			state.ObjAsVpcPeering().Status.State = string(code)
			changed = true
		}

		if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedAcceptingVpcPeeringConnection,
			Message: ptr.Deref(state.vpcPeering.Status.Message, ""),
		}) {
			changed = true
		}

		if changed {
			return composed.PatchStatus(state.ObjAsVpcPeering()).
				ErrorLogMessage("Error updating VpcPeering status while waiting for AWS VPC peering pending-acceptance").
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}

		return composed.StopAndForget, nil
	}

	return nil, nil
}
