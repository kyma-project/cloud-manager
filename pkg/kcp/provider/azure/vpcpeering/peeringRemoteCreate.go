package vpcpeering

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/kyma-project/cloud-manager/pkg/feature"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/utils/ptr"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func peeringRemoteCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	syncRequired := false

	if state.remotePeering != nil {

		if !feature.VpcPeeringSync.Value(ctx) {
			return nil, ctx
		}

		peeringSyncLevel := ptr.Deref(state.remotePeering.Properties.PeeringSyncLevel, "")
		if peeringSyncLevel == armnetwork.VirtualNetworkPeeringLevelFullyInSync ||
			peeringSyncLevel == armnetwork.VirtualNetworkPeeringLevelRemoteNotInSync {
			return nil, ctx
		}

		syncRequired = true

		logger.Info("Remote VPC peering sync required", "peeringSyncLevel", peeringSyncLevel)
	}

	// Allow gateway transit if remote gateway is used
	allowGatewayTransit := state.ObjAsVpcPeering().Spec.Details.UseRemoteGateway

	// params must be the same as in peeringRemoteLoad()
	err := state.remoteClient.CreateOrUpdatePeering(
		ctx,
		state.remoteNetworkId.ResourceGroup,
		state.remoteNetworkId.NetworkName(),
		state.ObjAsVpcPeering().Spec.Details.PeeringName,
		state.localNetworkId.String(),
		true,
		false,
		allowGatewayTransit,
	)

	if err == nil {
		if syncRequired {
			logger.Info("Remote VPC peering updated")
		} else {
			logger.Info("Remote VPC peering created")
		}

		return nil, ctx
	}

	if syncRequired {
		logger.Error(err, "Error updating remote VPC peering")
	} else {
		logger.Error(err, "Error creating remote VPC peering")
	}

	if azuremeta.IsTooManyRequests(err) {
		return composed.LogErrorAndReturn(err,
			"Too many requests on creating/updating remote VPC peering",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()),
			ctx,
		)
	}

	defaultMessage := "Error creating remote VPC peering"

	if syncRequired {
		defaultMessage = "Error updating remote VPC peering"
	}

	message, isWarning := azuremeta.GetErrorMessage(err, defaultMessage)

	if isWarning {
		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateWarning)
	} else {
		state.ObjAsVpcPeering().Status.State = string(cloudcontrolv1beta1.StateError)
	}

	condition := metav1.Condition{
		Type:    cloudcontrolv1beta1.ConditionTypeError,
		Status:  metav1.ConditionTrue,
		Reason:  cloudcontrolv1beta1.ReasonFailedCreatingVpcPeeringConnection,
		Message: message,
	}

	changed := meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady)

	if meta.SetStatusCondition(&state.ObjAsVpcPeering().Status.Conditions, condition) {
		changed = true
	}

	successError := composed.StopAndForget

	if !changed {
		return successError, ctx
	}

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		ErrorLogMessage("Error updating KCP VpcPeering status on failed create/update of remote VPC peering").
		SuccessError(successError).
		Run(ctx, state)

}
