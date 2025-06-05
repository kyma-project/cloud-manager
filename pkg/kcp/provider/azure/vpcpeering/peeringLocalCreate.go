package vpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func peeringLocalCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.localPeering != nil {
		return nil, nil
	}

	// params must be the same as in peeringLocalLoad()
	// CreatePeering requires that client is authorized to access local and remote subscription.
	// If remote and local subscriptions are on different tenants CreatePeering in local subscription will fail if SPN
	// does not exist in remote tenant.
	err := state.localPeeringClient.CreatePeering(
		ctx,
		state.localNetworkId.ResourceGroup,
		state.localNetworkId.NetworkName(),
		state.ObjAsVpcPeering().GetLocalPeeringName(),
		state.remoteNetworkId.String(),
		true,
		state.ObjAsVpcPeering().Spec.Details.UseRemoteGateway,
	)
	if err != nil {
		logger.Error(err, "Error creating VPC Peering")

		if azuremeta.IsTooManyRequests(err) {
			return composed.LogErrorAndReturn(err,
				"Too many requests on creating local VPC peering",
				composed.StopWithRequeueDelay(util.Timing.T60000ms()),
				ctx,
			)
		}

		message, isWarning := azuremeta.GetErrorMessage(err, "Error creating VPC peering")

		changed := false

		if isWarning {
			if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.StateWarning) {
				state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateWarning)
				changed = true
			}
		} else {
			if state.ObjAsVpcPeering().Status.State != string(cloudcontrolv1beta1.StateError) {
				state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)
				changed = true
			}
		}

		condition := metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
			Message: message,
		}

		if composed.AnyConditionChanged(state.ObjAsVpcPeering(), condition) {
			changed = true
		}

		if !changed {
			return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
		}

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(condition).
			ErrorLogMessage("Error updating KCP VpcPeering status on failed creation of local VPC peering").
			FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	logger.Info("Local VPC peering created")
	return nil, nil
}
