package network

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureMeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func resourceGroupLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	rg, err := state.azureClient.GetResourceGroup(ctx, state.resourceGroupName)
	if azureMeta.IgnoreNotFoundError(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading CloudManager Azure resource group", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	if err != nil {
		// RG not found
		logger.Info("Azure CloudManager resource group not found")
		return nil, nil
	}

	logger.Info("Azure CloudManager resource group loaded")

	state.resourceGroup = rg

	return nil, nil
}
