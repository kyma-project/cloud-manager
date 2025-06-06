package vpcpeering

import (
	"context"
	"k8s.io/apimachinery/pkg/api/meta"

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

	if err == nil {
		state.remoteVpc = network
		return nil, ctx
	}

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

	statusState := string(cloudcontrolv1beta1.StateError)

	if isWarning {
		statusState = string(cloudcontrolv1beta1.StateWarning)
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

	changed := false

	if state.ObjAsVpcPeering().Status.State != statusState {
		state.ObjAsVpcPeering().Status.State = statusState
		changed = true
	}

	if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), condition) {
		changed = true
	}

	if !changed {
		return successError, nil
	}

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		ErrorLogMessage("Error updating VpcPeering status due to failed loading of remote VPC network").
		SuccessError(successError).
		Run(ctx, state)

}
