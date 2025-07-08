package vpcpeering

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/meta"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func peeringRemoteLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// remote client is not created
	if state.remoteClient == nil {
		logger.Info("Azure remote client not initialized. Skipping loading of remote peering")
		return nil, nil
	}

	var resourceGroup, networkName string
	ok := false

	if len(state.ObjAsVpcPeering().Status.RemoteId) > 0 {
		resourceID, err := azureutil.ParseResourceID(state.ObjAsVpcPeering().Status.RemoteId)
		if err == nil {
			resourceGroup = resourceID.ResourceGroup
			networkName = resourceID.ResourceName
			ok = true
		}

	}

	if !ok && state.remoteNetworkId != nil {
		resourceGroup = state.remoteNetworkId.ResourceGroup
		networkName = state.remoteNetworkId.NetworkName()
		ok = true
	}

	if !ok {
		return nil, nil
	}

	// params must be the same as in peeringRemoteCreate()
	peering, err := state.remoteClient.GetPeering(
		ctx,
		resourceGroup,
		networkName,
		state.ObjAsVpcPeering().Spec.Details.PeeringName,
	)

	statusState := string(cloudcontrolv1beta1.StateError)
	message := "Error loading remote VPC peering"

	if err == nil {
		if ptr.Deref(peering.Properties.RemoteVirtualNetwork.ID, "") != state.localNetworkId.String() {
			err = fmt.Errorf("peering with the same name already %s exists in network %s", state.ObjAsVpcPeering().Spec.Details.PeeringName, state.localNetworkId.String())
			message = fmt.Sprintf("Peering with the same name %s already exists in network %s", state.ObjAsVpcPeering().Spec.Details.PeeringName, state.remoteNetworkId.String())
			statusState = string(cloudcontrolv1beta1.StateWarning)
		}
	}

	if err == nil {

		logger = logger.WithValues("remotePeeringId", ptr.Deref(peering.ID, ""))
		ctx = composed.LoggerIntoCtx(ctx, logger)

		state.remotePeering = peering

		logger.Info("Remote VPC peering loaded")

		return nil, ctx
	}

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return composed.LogErrorAndReturn(err, "Ignoring as marked for deletion", nil, ctx)
	}

	if azuremeta.IsNotFound(err) {
		return nil, nil
	}

	if azuremeta.IsTooManyRequests(err) {
		return composed.LogErrorAndReturn(err,
			"Too many requests on loading remote VPC peering",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()),
			ctx,
		)
	}

	logger.Error(err, "Error loading Azure remote VPC peering")

	message, isWarning := azuremeta.GetErrorMessage(err, message)

	if isWarning {
		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateWarning)
	}

	reason := cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcPeeringConnection
	successError := composed.StopAndForget

	if azuremeta.IsUnauthorized(err) {
		reason = cloudcontrolv1beta1.ReasonUnauthorized
		successError = composed.StopWithRequeueDelay(util.Timing.T300000ms())
	}

	if azuremeta.IsUnauthenticated(err) {
		successError = composed.StopWithRequeueDelay(util.Timing.T300000ms())
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

	if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
		changed = true
	}

	if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), condition) {
		changed = true
	}

	if !changed {
		return successError, nil
	}

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		ErrorLogMessage("Error updating KCP VpcPeering status on failed loading of remote VPC peering").
		SuccessError(successError).
		Run(ctx, state)
}
