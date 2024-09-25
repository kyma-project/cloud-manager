package network

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureMeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func vnetLoad(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	net, err := state.azureClient.GetNetwork(ctx, state.resourceGroupName, state.ObjAsNetwork().Name)
	if azureMeta.IgnoreNotFoundError(err) != nil {
		return composed.LogErrorAndReturn(err, "Error loading Azure vnet for KCP Network", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	if err != nil {
		// not found
		return nil, nil
	}

	logger.Info("Azure vnet loaded for KCP Network")

	state.network = net

	return nil, nil
}
