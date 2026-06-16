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

	// Azure may still report Succeeded briefly after the DELETE is issued before transitioning
	// to Deleting — requeue in all non-nil states until the link disappears.
	logger.Info("Azure virtual network link instance is still being deleted, requeueing with delay")
	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
