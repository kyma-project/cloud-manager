package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// deleteFileSystem deletes the NAS file system.
func deleteFileSystem(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.fileSystem == nil {
		return nil, ctx
	}

	logger.Info("Deleting AliCloud NAS file system", "fileSystemId", state.fileSystemId)

	if err := state.client.DeleteFileSystem(ctx, state.fileSystemId); err != nil {
		logger.Error(err, "Error deleting AliCloud NAS file system", "fileSystemId", state.fileSystemId)
		return composed.StopWithRequeue, ctx
	}

	return composed.StopWithRequeue, ctx
}
