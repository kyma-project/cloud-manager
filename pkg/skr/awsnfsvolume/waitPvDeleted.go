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

	logger.Info("waitPvDeleted: start")

	if state.Volume == nil {
		logger.Info("waitPvDeleted: no vol")
		return nil, nil
	}

	logger.Info("waitPvDeleted: wait for pv is deleted")

	// wait until PV does not exist
	return composed.StopWithRequeueDelay(250 * time.Millisecond), nil
}
