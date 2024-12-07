package nuke

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func checkIfAllProviderResourcesDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// check if some provider resources have been loaded to state, if none is loaded then they are all deleted
	allDeleted := true
	for _, prks := range state.ProviderResources {
		if len(prks.Objects) > 0 {
			allDeleted = false
			break
		}
	}

	if allDeleted {
		logger.Info("All orphan provided resources nuke deleted")

		return nil, ctx
	}

	logger.Info("Waiting for orphan provider resources to get nuke deleted")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
