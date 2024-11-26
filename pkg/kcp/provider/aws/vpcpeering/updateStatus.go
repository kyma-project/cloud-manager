package vpcpeering

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	cloudcontrol1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	obj := state.ObjAsVpcPeering()

	statusState := obj.Status.State

	if state.vpcPeering != nil {
		codes := map[ec2types.VpcPeeringConnectionStateReasonCode]string{
			ec2types.VpcPeeringConnectionStateReasonCodeInitiatingRequest: cloudcontrol1beta1.VpcPeeringConnectionStateReasonCodeInitiatingRequest,
			ec2types.VpcPeeringConnectionStateReasonCodePendingAcceptance: cloudcontrol1beta1.VpcPeeringConnectionStateReasonCodePendingAcceptance,
			ec2types.VpcPeeringConnectionStateReasonCodeActive:            cloudcontrol1beta1.VpcPeeringConnectionStateReasonCodeActive,
			ec2types.VpcPeeringConnectionStateReasonCodeDeleted:           cloudcontrol1beta1.VpcPeeringConnectionStateReasonCodeDeleted,
			ec2types.VpcPeeringConnectionStateReasonCodeRejected:          cloudcontrol1beta1.VpcPeeringConnectionStateReasonCodeRejected,
			ec2types.VpcPeeringConnectionStateReasonCodeFailed:            cloudcontrol1beta1.VpcPeeringConnectionStateReasonCodeFailed,
			ec2types.VpcPeeringConnectionStateReasonCodeExpired:           cloudcontrol1beta1.VpcPeeringConnectionStateReasonCodeExpired,
			ec2types.VpcPeeringConnectionStateReasonCodeProvisioning:      cloudcontrol1beta1.VpcPeeringConnectionStateReasonCodeProvisioning,
			ec2types.VpcPeeringConnectionStateReasonCodeDeleting:          cloudcontrol1beta1.VpcPeeringConnectionStateReasonCodeDeleting,
		}

		val, ok := codes[state.vpcPeering.Status.Code]
		if ok {
			statusState = val
		}
	}

	if len(obj.Status.Id) > 0 &&
		len(obj.Status.RemoteId) > 0 &&
		meta.IsStatusConditionTrue(*obj.Conditions(), cloudcontrol1beta1.ConditionTypeReady) &&
		obj.Status.State == statusState {
		// all already set and saved
		return nil, nil
	}

	obj.Status.State = statusState

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudcontrol1beta1.ConditionTypeReady,
			Status:  "True",
			Reason:  cloudcontrol1beta1.ReasonReady,
			Message: "Additional VpcPeerings(s) are provisioned",
		}).
		ErrorLogMessage("Error updating VpcPeering success status after setting Ready condition").
		SuccessLogMsg("KPC VpcPeering is ready").
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}
