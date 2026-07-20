package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// waitMountTargetsDeleted requeues until no NAS mount targets remain for the file system.
func waitMountTargetsDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.fileSystemId == "" {
		return nil, ctx
	}

	mts, err := state.client.DescribeMountTargets(ctx, state.fileSystemId)
	if err != nil {
		logger.Error(err, "Error describing AliCloud NAS mount targets while waiting for deletion")
		return composed.StopWithRequeue, ctx
	}

	state.mountTargets = mts

	if len(mts) > 0 {
		logger.Info("Waiting for AliCloud NAS mount targets to be deleted", "remaining", len(mts))
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}

	return nil, ctx
}
