package vnetlink

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
)

func createVNetLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.vnetLink != nil {
		return nil, ctx
	}

	kymaNetworkName := state.Scope().Spec.Scope.Azure.VpcNetwork
	kymaResourceGroup := state.Scope().Spec.Scope.Azure.VpcNetwork
	vnetId := azureutil.NewVirtualNetworkResourceId(state.Scope().Spec.Scope.Azure.SubscriptionId,
		kymaResourceGroup, kymaNetworkName).String()

	err := state.remoteClient.CreateVirtualNetworkLink(ctx,
		state.remotePrivateDnsZoneId.ResourceGroup,
		state.remotePrivateDnsZoneId.ResourceName,
		state.ObjAsAzureVNetLink().Spec.RemoteVirtualPrivateLinkName,
		vnetId)

	if err == nil {
		logger.Info("VirtualNetworkLink created")
		return nil, ctx
	}

	logger.Error(err, "Error creating VirtualNetworkLink")

	return azuremeta.HandleError(err, state.ObjAsAzureVNetLink()).
		WithDefaultReason(cloudcontrolv1beta1.ReasonFailedCreatingVirtualNetworkLink).
		WithDefaultMessage("Failed creating VirtualNetworkLink").
		WithTooManyRequestsMessage("Too many requests on creating VirtualNetworkLink").
		WithUpdateStatusMessage("Error updating KCP AzureVNetLink status on failed creating of VirtualNetworkLink").
		Run(ctx, state)

}
