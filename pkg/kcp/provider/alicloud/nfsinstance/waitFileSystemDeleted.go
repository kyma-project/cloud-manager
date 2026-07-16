package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
)

// waitFileSystemDeleted requeues until the NAS file system no longer exists.
func waitFileSystemDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.fileSystemId == "" {
		return nil, ctx
	}

	fs, err := state.client.DescribeFileSystem(ctx, state.fileSystemId)
	if err != nil {
		logger.Error(err, "Error describing AliCloud NAS file system while waiting for deletion")
		return composed.StopWithRequeue, ctx
	}

	if fs != nil {
		logger.Info("Waiting for AliCloud NAS file system to be deleted", "fileSystemId", state.fileSystemId)
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx
	}

	state.fileSystem = nil

	return nil, ctx
}
