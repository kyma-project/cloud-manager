package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func resourceGroupWatcherCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.resourceGroupWatcher != nil {
		return nil, ctx
	}

	logger.Info("Creating NetworkWatcherRG")
	rg, err := state.azureClient.CreateResourceGroup(ctx,
		state.resourceGroupWatcherName(),
		state.location(),
		nil)
	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error creating NetworkWatcherRG", ctx)
	}

	state.resourceGroupWatcher = rg
	return nil, ctx
}
