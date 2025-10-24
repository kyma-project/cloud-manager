package vpcpeering

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
)

func initLocalNetworkId(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	state.localNetworkId = azureutil.NewVirtualNetworkResourceIdFromNetworkReference(state.LocalNetwork().Status.Network)

	logger := composed.LoggerFromCtx(ctx)

	logger.WithValues(
		"localNetworkAzureId", state.localNetworkId.String(),
	)

	logger.Info("KCP VpcPeering local network id initialized")

	return nil, composed.LoggerIntoCtx(ctx, logger)
}
