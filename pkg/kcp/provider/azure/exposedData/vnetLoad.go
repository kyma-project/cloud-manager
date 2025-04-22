package exposedData

import (
	"context"
	"errors"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func vnetLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.networkId == nil || !state.networkId.IsValid() {
		err := errors.New("logical error")
		return composed.LogErrorAndReturn(err, "KCP Kyma Network has invalid network id", composed.StopAndForget, ctx)
	}

	vnet, err := state.azureClient.GetNetwork(ctx, state.networkId.ResourceName, state.networkId.NetworkName())
	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading Azure vnet", ctx)
	}

	state.vnet = vnet

	return nil, ctx
}
