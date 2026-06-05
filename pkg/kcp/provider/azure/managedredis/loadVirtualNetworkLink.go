package managedredis

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// virtualNetworkLinkName returns the shared per-shoot link name. All AzureManagedRedis
// resources in the same shoot reuse a single link from the privatelink.redis.azure.net
// zone (in the Cloud Manager resource group) to the shoot's gardener-managed vnet, so
// that pods running in worker nodes resolve the AMR FQDN to its private endpoint IP.
// Naming the link after the gardener vnet makes the create-if-missing path idempotent
// across siblings.
func virtualNetworkLinkName(state *State) string {
	return state.gardenerNetworkName
}

func loadVirtualNetworkLink(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	link, err := state.client.GetVirtualNetworkLink(ctx, state.resourceGroupName, state.PrivateDNSZoneName(), virtualNetworkLinkName(state))
	if err != nil {
		if azuremeta.IsNotFound(err) {
			return nil, ctx
		}
		return composed.LogErrorAndReturn(err, "Error loading Azure Managed Redis virtual network link", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	state.virtualNetworkLink = link
	return nil, ctx
}
