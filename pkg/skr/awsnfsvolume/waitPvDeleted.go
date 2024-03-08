package awsnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func waitPvDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.Volume == nil {
		logger.Info("PersistentVolume is deleted")
		return nil, nil
	}

	logger.Info("Waiting for PersistentVolume to be deleted")

	// wait until PV does not exist / gets deleted
	return composed.StopWithRequeueDelay(250 * time.Millisecond), nil
}
