package vpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func vpcRemoteLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.remotePeering != nil {
		return nil, nil
	}

	network, err := state.remoteClient.GetNetwork(ctx, state.remoteNetworkId.ResourceGroup, state.remoteNetworkId.NetworkName())

	if err != nil {
		logger.Error(err, "Error loading remote network")

		message, isWarning := azuremeta.GetErrorMessage(err, "Error loading remote network")

		successError := composed.StopWithRequeueDelay(util.Timing.T60000ms())

		// If VpcNetwork is not found user can not recover from this error without updating the resource so, we are doing
		// stop and forget.
		if azuremeta.IsNotFound(err) {
			successError = composed.StopAndForget
			message = "Remote VPC network not found"
			logger.Info(message)
		}

		if isWarning {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateWarning)
		} else {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)
		}

		reason := cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcNetwork

		if azuremeta.IsUnauthorized(err) {
			reason = cloudcontrolv1beta1.ReasonUnauthorized
		}

		condition := metav1.Condition{
			Type:    cloudcontrolv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  reason,
			Message: message,
		}

		if !composed.AnyConditionChanged(state.ObjAsVpcPeering(), condition) {
			return successError, nil
		}

		return composed.UpdateStatus(state.ObjAsVpcPeering()).
			SetCondition(condition).
			ErrorLogMessage("Error updating VpcPeering status due to failed loading of remote VPC network").
			FailedError(composed.StopWithRequeue).
			SuccessError(successError).
			Run(ctx, state)
	}

	state.remoteVpc = network

	return nil, nil
}
