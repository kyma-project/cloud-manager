package vnetlink

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func loadPrivateDnsZone(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	privateDnsZone, err := state.remoteClient.GetPrivateDnsZone(ctx,
		state.remotePrivateDnsZoneId.ResourceGroup,
		state.remotePrivateDnsZoneId.ResourceName)

	state.privateDnzZone = privateDnsZone

	return azuremeta.HandleLoadingError("PrivateDnsZone", err, ctx)
}
