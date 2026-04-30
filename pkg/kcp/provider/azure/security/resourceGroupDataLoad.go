package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func resourceGroupDataLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	rg, err := state.azureClient.GetResourceGroup(ctx, state.resourceGroupDataName())
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading security data resource group", ctx)
	}
	if err == nil {
		state.resourceGroupData = rg
	}

	return nil, ctx
}
