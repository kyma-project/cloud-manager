package sapnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func pvWaitDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}

	if state.PV == nil {
		return nil, ctx
	}

	logger.Info("Waiting for SapNfsVolume PV to be deleted")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
