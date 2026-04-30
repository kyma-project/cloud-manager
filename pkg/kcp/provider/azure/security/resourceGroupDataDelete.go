package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func resourceGroupDataDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroupData == nil {
		return nil, ctx
	}

	logger.Info("Deleting security resource group")
	err := state.azureClient.DeleteResourceGroup(ctx, state.resourceGroupDataName())
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return azuremeta.LogErrorAndReturn(err, "Error deleting security resource group", ctx)
	}

	state.resourceGroupData = nil
	return nil, ctx
}
