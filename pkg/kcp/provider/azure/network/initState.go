package network

import (
	"context"
	"fmt"
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

	// Tags
	state.tags = map[string]string{
		common.TagCloudManagerName: state.Name().String(),
		common.TagShoot:            state.Scope().Spec.ShootName,
		common.TagScope:            fmt.Sprintf("%s/%s", state.Scope().Namespace, state.Scope().Name),
	}

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
