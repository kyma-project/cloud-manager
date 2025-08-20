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

func peeringLocalCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	syncRequired := false
	if state.localPeering != nil {

		if !feature.VpcPeeringSync.Value(ctx) {
			return nil, ctx
		}

		peeringSyncLevel := ptr.Deref(state.localPeering.Properties.PeeringSyncLevel, "")
		if peeringSyncLevel == armnetwork.VirtualNetworkPeeringLevelFullyInSync ||
			peeringSyncLevel == armnetwork.VirtualNetworkPeeringLevelRemoteNotInSync {
			return nil, ctx
		}

		syncRequired = true
		logger.Info("Local VPC peering sync required", "peeringSyncLevel", peeringSyncLevel)
	}

	// params must be the same as in peeringLocalLoad()
	// CreateOrUpdatePeering requires that client is authorized to access local and remote subscription.
	// If remote and local subscriptions are on different tenants CreateOrUpdatePeering in local subscription will fail if SPN
	// does not exist in remote tenant.
	err := state.localPeeringClient.CreateOrUpdatePeering(
		ctx,
		state.localNetworkId.ResourceGroup,
		state.localNetworkId.NetworkName(),
		state.ObjAsVpcPeering().GetLocalPeeringName(),
		state.remoteNetworkId.String(),
		true,
		state.ObjAsVpcPeering().Spec.Details.UseRemoteGateway,
		false,
	)

	if err == nil {
		if syncRequired {
			logger.Info("Local VPC peering updated")
		} else {
			logger.Info("Local VPC peering created")
		}

		return nil, ctx
	}

	if syncRequired {
		logger.Error(err, "Error updating VPC Peering")
	} else {
		logger.Error(err, "Error creating VPC Peering")
	}

	if azuremeta.IsTooManyRequests(err) {
		return composed.LogErrorAndReturn(err,
			"Too many requests on creating/updating local VPC peering",
			composed.StopWithRequeueDelay(util.Timing.T60000ms()),
			ctx,
		)
	}

	defaultMassage := "Error creating VPC peering"

	if syncRequired {
		defaultMassage = "Error updating VPC peering"
	}

	message, isWarning := azuremeta.GetErrorMessage(err, defaultMassage)

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

	if meta.RemoveStatusCondition(state.ObjAsVpcPeering().Conditions(), cloudcontrolv1beta1.ConditionTypeReady) {
		changed = true
	}

	if meta.SetStatusCondition(state.ObjAsVpcPeering().Conditions(), condition) {
		changed = true
	}

	successError := composed.StopAndForget

	if !changed {
		return successError, ctx
	}

	return composed.PatchStatus(state.ObjAsVpcPeering()).
		ErrorLogMessage("Error updating KCP VpcPeering status on failed create/update of local VPC peering").
		SuccessError(successError).
		Run(ctx, state)

}
