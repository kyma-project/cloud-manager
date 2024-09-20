package iprange

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func kymaPeeringDeleteWait(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	if state.kymaPeering == nil {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Waiting KCP VpcPeering for IpRange to be deleted")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
