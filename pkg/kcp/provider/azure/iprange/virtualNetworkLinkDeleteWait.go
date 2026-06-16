package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func virtualNetworkLinkDeleteWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.virtualNetworkLink == nil {
		return nil, nil
	}

	// Resource keeps prior provisioningState until ARM moves it to Deleting; requeue until gone.
	// https://learn.microsoft.com/en-us/azure/azure-resource-manager/management/async-operations#provisioningstate-values
	logger.Info("Azure virtual network link instance is still being deleted, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
