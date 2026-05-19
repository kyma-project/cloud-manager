package security

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func resourceGroupWatcherLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	rg, err := state.azureClient.GetResourceGroup(ctx, state.resourceGroupWatcherName())
	if azuremeta.IgnoreNotFoundError(err) != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error loading watcher resource group: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error loading NetworkWatcherRG", ctx)
	}
	if err == nil {
		state.resourceGroupWatcher = rg
	}

	return nil, ctx
}
