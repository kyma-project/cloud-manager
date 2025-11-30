package scope

import (
	"context"

	commongardener "github.com/kyma-project/cloud-manager/pkg/common/gardener"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	scopeconfig "github.com/kyma-project/cloud-manager/pkg/kcp/scope/config"
)

func gardenerClientCreate(ctx context.Context, st composed.State) (error, context.Context) {
	logger := composed.LoggerFromCtx(ctx)
	state := st.(*State)

	logger.Info("Loading gardener credentials")

	out, err := commongardener.CreateGardenerClient(ctx, commongardener.CreateGardenerClientInput{
		KcpClient:                 state.Cluster().ApiReader(),
		GardenerFallbackNamespace: scopeconfig.ScopeConfig.GardenerNamespace,
	})

	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating gardener client", composed.StopWithRequeue, ctx)
	}

	state.shootNamespace = out.Namespace

	logger = logger.WithValues("shootNamespace", state.shootNamespace)
	logger.Info("Detected shoot namespace")

	state.gardenerClient = out.Client

	logger.Info("Gardener clients created")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
