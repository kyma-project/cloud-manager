package awsnfsvolume

import (
	"context"
	"time"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func waitPvDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, ctx
	}

	if state.Volume == nil {
		return nil, ctx
	}

	logger.Info("Waiting for PersistentVolume to be deleted")

	// wait until PV does not exist / gets deleted
	return composed.StopWithRequeueDelay(250 * time.Millisecond), ctx
}
