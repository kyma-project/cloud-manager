package vpcpeering

import (
	"context"
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

	// remote client not created
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

	if err != nil {
		if composed.MarkedForDeletionPredicate(ctx, state) {
			return composed.LogErrorAndReturn(err, "Ignoring as marked for deletion", nil, ctx)
		}

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

		logger.Error(err, "Error loading remote VPC Peering")

		message, isWarning := azuremeta.GetErrorMessage(err)

		if isWarning {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.WarningState)
		} else {
			state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.ErrorState)
		}

		reason := cloudcontrolv1beta1.ReasonFailedLoadingRemoteVpcPeeringConnection

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
			return nil, nil
		}

		return composed.PatchStatus(state.ObjAsVpcPeering()).
			SetExclusiveConditions(condition).
			ErrorLogMessage("Error updating KCP VpcPeering status on failed loading of remote vpc peering").
			FailedError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).
			SuccessErrorNil().
			Run(ctx, state)
	}

	logger = logger.WithValues("remotePeeringId", ptr.Deref(peering.ID, ""))
	ctx = composed.LoggerIntoCtx(ctx, logger)

	state.remotePeering = peering

	logger.Info("Azure remote VPC peering loaded")

	return nil, ctx
}
