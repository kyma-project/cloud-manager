package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// loadMountTargets loads the NAS mount targets for the file system into state.
func loadMountTargets(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.fileSystemId == "" {
		return nil, ctx
	}

	mts, err := state.client.DescribeMountTargets(ctx, state.fileSystemId)
	if err != nil {
		logger.Error(err, "Error loading AliCloud NAS mount targets", "fileSystemId", state.fileSystemId)
		return composed.StopWithRequeue, ctx
	}

	state.mountTargets = mts

	return nil, ctx
}
