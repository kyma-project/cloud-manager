package security

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
)

func networkWatcherCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.watcher != nil {
		return nil, ctx
	}
	if state.resourceGroupWatcher == nil {
		return composed.LogErrorAndReturn(
			fmt.Errorf("resourceGroupWatcher must exist before creating network watcher"),
			"Cannot create network watcher",
			composed.StopWithRequeue, ctx)
	}

	watcher, err := state.azureClient.CreateNetworkWatcher(ctx,
		ResourceGroupWatcherName(),
		NetworkWatcherName(state.ObjAsRuntime().Spec.Shoot.Region),
		armnetwork.Watcher{
			Location: new(state.ObjAsRuntime().Spec.Shoot.Region),
		})
	if err != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error creating network watcher: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error creating network watcher", ctx)
	}

	state.watcher = watcher

	return nil, ctx
}
