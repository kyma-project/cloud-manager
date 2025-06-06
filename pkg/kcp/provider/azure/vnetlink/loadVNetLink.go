package vnetlink

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func loadVNetLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	vnetLink, err := state.remoteClient.GetVirtualNetworkLink(ctx,
		state.remotePrivateDnsZoneId.ResourceGroup,
		state.remotePrivateDnsZoneId.ResourceName,
		state.ObjAsAzureVNetLink().Spec.RemoteVNetLinkName)

	if err == nil {
		ctx = composed.LoggerIntoCtx(ctx, logger.WithValues("vnetLinkId", ptr.Deref(vnetLink.ID, "")))
		state.vnetLink = vnetLink
		return nil, ctx
	}

	return azuremeta.HandleError(err, state.ObjAsAzureVNetLink()).
		WithDefaultReason(cloudcontrolv1beta1.ReasonFailedCreatingVirtualNetworkLink).
		WithDefaultMessage("Failed loading VirtualNetworkLink").
		WithTooManyRequestsMessage("Too many requests on loading VirtualNetworkLink").
		WithUpdateStatusMessage("Error updating KCP AzureVNetLink status on failed loading of VirtualNetworkLink").
		WithNotFoundMessage("VirtualNetworkLink not found").
		Run(ctx, state)
}
