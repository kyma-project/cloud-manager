package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func peeringRemoteLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// remote client not created
	if state.remoteClient == nil {
		return nil, nil
	}

	// params must be the same as in peeringRemoteCreate()
	peering, err := state.remoteClient.GetPeering(
		ctx,
		state.remoteNetworkId.ResourceGroup,
		state.remoteNetworkId.NetworkName(),
		state.ObjAsVpcPeering().Spec.Details.PeeringName,
	)
	if azuremeta.IsNotFound(err) {
		return nil, nil
	}
	if azuremeta.IsTooManyRequests(err) {
		return composed.LogErrorAndReturn(err,
			"Azure vpc peering too many requests on peering remote load",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()),
			ctx,
		)
	}

	if err != nil {
		logger.Error(err, "Error loading remote VPC Peering")

		message, isWarning := azuremeta.GetErrorMessage(err)

		if isWarning {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.WarningState)
		} else {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.ErrorState)
		}

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
				Message: message,
			}).
			ErrorLogMessage("Error updating KCP VpcPeering status on failed loading of remote vpc peering").
			FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).
			Run(ctx, state)
	}

	logger = logger.WithValues("remotePeeringId", ptr.Deref(peering.ID, ""))
	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.remotePeering = peering

	logger.Info("Azure remote VPC peering loaded")

	return nil, ctx
}
