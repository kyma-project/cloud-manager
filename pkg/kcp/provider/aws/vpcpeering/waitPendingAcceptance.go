package vpcpeering

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
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
	if awsutil.IsTerminatedOrDeleting(state.vpcPeering) {
		changed := false

		if state.ObjAsVpcPeering().Status.State != string(code) {
			state.ObjAsVpcPeering().Status.State = string(code)
			changed = true
		}

		if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
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
				SuccessLogMsg(fmt.Sprintf("VpcPeering status %s updated", code)).
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}

		return composed.StopAndForget, ctx
	}

	return nil, ctx
}
