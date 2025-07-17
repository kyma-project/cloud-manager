package sapnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func pvcWaitDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}

	if state.PVC == nil {
		return nil, ctx
	}

	logger.Info("Waiting for SapNfsVolume PVC to be deleted")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
