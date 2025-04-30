package exposedData

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func subnetsLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.vnet == nil {
		return composed.LogErrorAndReturn(common.ErrLogical, "vnet should be defined", composed.StopAndForget, ctx)
	}

	for _, subnetRef := range state.vnet.Properties.Subnets {
		// subnet must be in the same resource group as the vnet
		subnet, err := state.azureClient.GetSubnet(ctx, state.networkId.ResourceGroup, state.networkId.NetworkName(), ptr.Deref(subnetRef.Name, ""))
		if err != nil {
			return azuremeta.LogErrorAndReturn(err, "Error loading Subnet", ctx)
		}
		state.subnets = append(state.subnets, subnet)
	}

	return nil, ctx
}
