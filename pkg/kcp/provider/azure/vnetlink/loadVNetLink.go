package vnetlink

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func loadVNetLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	vnetLink, err := state.remoteClient.GetVirtualNetworkLink(ctx,
		state.remotePrivateDnsZoneId.ResourceGroup,
		state.remotePrivateDnsZoneId.ResourceName,
		state.ObjAsAzureVNetLink().Spec.RemoteVirtualPrivateLinkName)

	if vnetLink != nil {
		logger := log.FromContext(ctx)
		ctx = composed.LoggerIntoCtx(ctx, logger.WithValues("vnetLinkId", ptr.Deref(vnetLink.ID, "")))

		state.vnetLink = vnetLink
	}

	return azuremeta.HandleLoadingError("VirtualNetworkLink", err, ctx)
}
