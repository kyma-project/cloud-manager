package nuke

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func checkIfAllDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// check if some resources have been loaded to state, if none is loaded then they are all deleted
	allDeleted := true
	for _, rks := range state.Resources {
		if len(rks.Objects) > 0 {
			allDeleted = false
			break
		}
	}

	if allDeleted {
		logger.Info("All orphan resources nuke deleted")

		return nil, ctx
	}

	logger.Info("Waiting for orphan resources to get nuke deleted")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
