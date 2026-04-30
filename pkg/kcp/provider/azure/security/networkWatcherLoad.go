package security

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"k8s.io/utils/ptr"
)

func networkWatcherLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	list, err := state.azureClient.ListNetworkWatchers(ctx)
	if err != nil {
		return azuremeta.LogErrorAndReturn(err, "Error listing network watchers", ctx)
	}

	for _, w := range list {
		if ptr.Deref(w.Location, "") == state.location() {
			state.watcher = w
			break
		}
	}

	return nil, ctx
}
