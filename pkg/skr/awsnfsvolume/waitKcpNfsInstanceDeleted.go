package awsnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func waitKcpNfsInstanceDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	if state.KcpNfsInstance == nil {
		logger.Info("Kcp NfsInstance is deleted")
		return nil, nil
	}

	logger.Info("Waiting for Kcp NfsInstance to be deleted")

	// TODO: check if KCP instance got some Error condition related to deletion
	// in that case we have to StopAndForget and not to keep looping

	// wait until KcpNfsInstance does not exist / gets deleted
	return composed.StopWithRequeueDelay(250 * time.Millisecond), nil
}
