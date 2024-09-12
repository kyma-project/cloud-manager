package network

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func resourceGroupDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroup == nil {
		return nil, nil
	}

	logger.Info("Deleting Azure resource group for KCP Network")

	err := state.azureClient.DeleteResourceGroup(ctx, state.resourceGroupName)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting Azure resource group for KCP Network", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return nil, nil
}
