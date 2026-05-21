package security

import (
	"context"
	"fmt"

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
		ResourceGroupWatcherName(),
		state.ObjAsRuntime().Spec.Shoot.Region,
		nil)
	if err != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error creating watcher resource group: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error creating NetworkWatcherRG", ctx)
	}

	state.resourceGroupWatcher = rg

	return nil, ctx
}
