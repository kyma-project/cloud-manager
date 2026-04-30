package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func resourceGroupWatcherLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	rg, err := state.azureClient.GetResourceGroup(ctx, state.resourceGroupWatcherName())
	if azuremeta.IgnoreNotFoundError(err) != nil {
		return azuremeta.LogErrorAndReturn(err, "Error loading NetworkWatcherRG", ctx)
	}
	if err == nil {
		state.resourceGroupWatcher = rg
	}

	return nil, ctx
}
