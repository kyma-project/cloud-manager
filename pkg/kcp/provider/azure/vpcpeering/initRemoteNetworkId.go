package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
)

func initRemoteNetworkId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if state.RemoteNetwork() == nil {
		return nil, ctx
	}

	state.remoteNetworkId = azureutil.NewVirtualNetworkResourceIdFromNetworkReference(state.RemoteNetwork().Status.Network)

	logger := composed.LoggerFromCtx(ctx)

	logger.WithValues(
		"remoteNetwork", state.RemoteNetwork().Name,
		"remoteNetworkAzureId", state.remoteNetworkId.String(),
	)

	logger.Info("KCP VpcPeering remote network id initialized")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
