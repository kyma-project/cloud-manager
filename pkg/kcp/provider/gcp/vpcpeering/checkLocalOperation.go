package vpcpeering

import (
	"context"
	"strings"

	pb "cloud.google.com/go/compute/apiv1/computepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkLocalOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if composed.MarkedForDeletionPredicate(ctx, state) ||
		state.ObjAsVpcPeering().Status.Id != "" ||
		state.localPeeringOperation != nil ||
		state.ObjAsVpcPeering().Status.LocalPeeringOperation == "" {
		return nil, ctx
	}

	op, err := state.client.GetOperation(ctx, state.LocalNetwork().Spec.Network.Reference.Gcp.GcpProject, state.ObjAsVpcPeering().Status.LocalPeeringOperation)
	if err != nil {
		logger.Error(err, "[KCP GCP VpcPeering checkLocalOperation] Error getting local operation")
		meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), metav1.Condition{
			Type:    v1beta1.ConditionTypeError,
			Status:  "True",
			Reason:  v1beta1.ConditionReasonError,
			Message: "Error loading Local Vpc Peering LocalOperation: " + state.ObjAsVpcPeering().Status.LocalPeeringOperation,
		})
		err = state.PatchObjStatus(ctx)
		if err != nil {
			return composed.LogErrorAndReturn(err,
				"Error updating status since it was not possible to load the local Vpc Peering operation.",
				composed.StopWithRequeueDelay(util.Timing.T10000ms()),
				ctx,
			)
		}
		return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
	}
	state.localPeeringOperation = op
	if op != nil {
		if op.GetStatus() != pb.Operation_DONE {
			logger.Info("Local operation still in progress", "localPeeringOperation", ptr.Deref(op.Name, "OperationUnknown"))
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
		}
		if op.GetError() != nil {
			logger.Error(err, "Local operation error ", "localPeeringOperation", op.GetName())
			state.ObjAsVpcPeering().Status.State = cloudcontrolv1beta1.VirtualNetworkPeeringStateDisconnected
			if strings.Contains(op.GetError().String(), "QUOTA_EXCEEDED") {
				return composed.UpdateStatus(state.ObjAsVpcPeering()).SetExclusiveConditions(metav1.Condition{
					Type:    v1beta1.ConditionTypeQuotaExceeded,
					Status:  "True",
					Reason:  v1beta1.ConditionTypeQuotaExceeded,
					Message: "Error creating Local Vpc Peering due to quota limits, please check if your vpc quota limits are not exceeded.",
				}).
					ErrorLogMessage("Error creating Local VpcPeering due to quota exceeded").
					FailedError(composed.StopWithRequeue).
					SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
					Run(ctx, state)
			}

			return composed.UpdateStatus(state.ObjAsVpcPeering()).SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  "True",
				Reason:  v1beta1.ConditionTypeError,
				Message: "The cloud provider had an error while creating Local Vpc Peering",
			}).
				ErrorLogMessage("The cloud provider had an error while creating Local Vpc Peering").
				FailedError(composed.StopWithRequeue).
				SuccessError(composed.StopWithRequeueDelay(util.Timing.T300000ms())).
				Run(ctx, state)
		}
	}

	return nil, nil
}
