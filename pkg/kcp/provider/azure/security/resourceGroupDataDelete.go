package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func resourceGroupDataDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroupData == nil {
		return nil, ctx
	}

	logger.Info("Deleting security resource group")
	_, err := azureclient.PollUntilDone(state.azureClient.DeleteResourceGroup(ctx,
		state.resourceGroupDataName()))(ctx, nil)
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return azuremeta.LogErrorAndReturn(err, "Error deleting security resource group", ctx)
	}

	state.resourceGroupData = nil

	return nil, ctx
}
