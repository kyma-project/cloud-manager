package network

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
)

func initState(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// Resource Group Name
	state.resourceGroupName = azurecommon.AzureCloudManagerResourceGroupName(state.Scope().Spec.Scope.Azure.VpcNetwork)

	// Location
	state.location = state.ObjAsNetwork().Spec.Network.Managed.Location
	if state.location == "" {
		state.location = state.Scope().Spec.Region
	}

	// Tags, nil until we determine the use case justifying their presence and unify on tag names across providers
	state.tags = nil

	// Cidr
	state.cidr = state.ObjAsNetwork().Spec.Network.Managed.Cidr
	if len(state.cidr) == 0 {
		state.cidr = common.DefaultCloudManagerCidr
	}

	// Logger
	logger = logger.WithValues(
		"azureTenantId", state.Scope().Spec.Scope.Azure.TenantId,
		"azureSubscriptionId", state.Scope().Spec.Scope.Azure.SubscriptionId,
		"azureResourceGroup", state.resourceGroupName,
		"azureVirtualNetwork", state.ObjAsNetwork().Name,
		"azureVNetLocation", state.location,
		"azureVNetAddressSpace", state.cidr,
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return nil, ctx
}
