package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// waitFileSystemAvailable polls the NAS file system until it reaches the Running state.
func waitFileSystemAvailable(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.fileSystemId == "" {
		return composed.StopWithRequeue, ctx
	}

	fs, err := state.client.DescribeFileSystem(ctx, state.fileSystemId)
	if err != nil {
		logger.Error(err, "Error describing AliCloud NAS file system while waiting for availability")
		return composed.StopWithRequeue, ctx
	}

	if fs == nil || fs.Status != "Running" {
		logger.Info("Waiting for AliCloud NAS file system to become available", "fileSystemId", state.fileSystemId)
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}

	state.fileSystem = fs

	return nil, ctx
}
