package security

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func resourceGroupDataLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	rg, err := state.azureClient.GetResourceGroup(ctx, state.resourceGroupDataName())
	if azuremeta.IgnoreNotFoundError(err) != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error loading data resource group: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error loading security data resource group", ctx)
	}
	if err == nil {
		state.resourceGroupData = rg
	}

	return nil, ctx
}
