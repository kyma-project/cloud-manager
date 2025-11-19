package awsnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func waitPVCDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.PVC == nil {
		return nil, ctx
	}

	logger.Info("Waiting for PersistentVolumeClaim to be deleted")

	// wait until PVC does not exist / gets deleted
	return composed.StopWithRequeueDelay(2 * util.Timing.T100ms()), nil
}
