package security

import (
	"context"
	"fmt"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func networkWatcherLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	list, err := state.azureClient.ListNetworkWatchers(ctx)
	if err != nil {
		_, _ = state.PatchStatusAnnotations(ctx, "Error", fmt.Sprintf("Error loading network watcher: %s", err.Error()), state.ObjAsRuntime().Generation)
		return azuremeta.LogErrorAndReturn(err, "Error listing network watchers", ctx)
	}

	for _, w := range list {
		if ptr.Deref(w.Location, "") == state.ObjAsRuntime().Spec.Shoot.Region {
			state.watcher = w
			break
		}
	}

	return nil, ctx
}
