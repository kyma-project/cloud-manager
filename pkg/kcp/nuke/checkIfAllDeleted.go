package nuke

import (
	"context"
	"strings"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

func checkIfAllDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// check if some resources have been loaded to state, if none is loaded then they are all deleted
	var waitingInfo strings.Builder
	allDeleted := true
	for j, rks := range state.Resources {
		if len(rks.Objects) > 0 {
			allDeleted = false
			if j > 0 {
				waitingInfo.WriteString(";")
			}
			waitingInfo.WriteString(rks.Kind)
			waitingInfo.WriteString("(")
			for i, o := range rks.Objects {
				if i > 0 {
					waitingInfo.WriteString(",")
				}
				waitingInfo.WriteString(o.GetName())
			}
			waitingInfo.WriteString(")")
		}
	}

	if allDeleted {
		logger.Info("All orphan resources nuke deleted")

		return nil, ctx
	}

	logger.WithValues("waitingFor", waitingInfo.String()).Info("Waiting for orphan resources to get nuke deleted")

	return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
}
