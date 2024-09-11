package network

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func vnetDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.network == nil {
		return nil, nil
	}

	logger.Info("Deleting Azure VNet for KCP Network")

	err := state.azureClient.DeleteNetwork(ctx, state.resourceGroupName, state.ObjAsNetwork().Name)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting Azure VNet for KCP Network", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}

	return nil, nil
}
