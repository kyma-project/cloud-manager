package network

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func resourceGroupDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroup == nil {
		return nil, nil
	}

	logger.Info("Deleting Azure resource group for KCP Network")

	_, err := azureclient.PollUntilDone(state.azureClient.DeleteResourceGroup(ctx, state.resourceGroupName))(ctx, nil)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting Azure resource group for KCP Network", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return nil, nil
}
