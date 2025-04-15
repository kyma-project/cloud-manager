package exposedData

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
)

func vnetLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.networkId = azureutil.NewVirtualNetworkResourceIdFromNetworkReference(state.kcpNetwork.Status.Network)
	vnet, err := state.azureClient.GetNetwork(ctx, state.networkId.ResourceName, state.networkId.NetworkName())
	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading Azure vnet", ctx)
	}

	state.vnet = vnet

	return nil, ctx
}
