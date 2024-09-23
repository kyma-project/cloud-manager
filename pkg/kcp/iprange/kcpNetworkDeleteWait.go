package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func kcpNetworkDeleteWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	if state.network == nil {
		return nil, ctx
	}

	if !state.isCloudManagerNetwork {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Waiting KCP CM Network to be deleted")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
