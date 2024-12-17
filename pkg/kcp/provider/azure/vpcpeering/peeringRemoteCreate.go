package vpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func peeringRemoteCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.remotePeering != nil {
		return nil, nil
	}

	// params must be the same as in peeringRemoteLoad()
	err := state.remoteClient.CreatePeering(
		ctx,
		state.remoteNetworkId.ResourceGroup,
		state.remoteNetworkId.NetworkName(),
		state.ObjAsVpcPeering().Spec.Details.PeeringName,
		state.localNetworkId.String(),
		true,
	)

	if err != nil {
		logger.Error(err, "Error creating remote VPC peering")

		if azuremeta.IsTooManyRequests(err) {
			return composed.LogErrorAndReturn(err,
				"Too many requests on creating remote VPC peering",
				composed.StopWithRequeueDelay(util.Timing.T60000ms()),
				ctx,
			)
		}

		message, isWarning := azuremeta.GetErrorMessage(err)

		if isWarning {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateWarning)
		} else {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)
		}

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: message,
			}).
			ErrorLogMessage("Error updating KCP VpcPeering status on failed creation of remote VPC peering").
			FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	logger.Info("Remote VPC peering created")

	return nil, nil
}
