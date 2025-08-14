package subscription

import (
	"context"

	commongardener "github.com/kyma-project/cloud-manager/pkg/common/gardener"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/scope"
)

func gardenerClientCreate(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	out, err := commongardener.CreateGardenerClient(ctx, commongardener.CreateGardenerClientInput{
		KcpClient:                 state.Cluster().ApiReader(),
		GardenerFallbackNamespace: scope.ScopeConfig.GardenerNamespace,
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating Gardener client for Subscription", composed.StopWithRequeue, ctx)
	}

	state.gardenNamespace = out.Namespace
	state.gardenerClient = out.GardenerClient
	state.gardenK8sClient = out.GardenK8sClient

	return nil, ctx
}
