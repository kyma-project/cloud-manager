package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

// loadFileSystem loads the NAS file system referenced by Status.Id into state, if it exists.
func loadFileSystem(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	fsId := state.ObjAsNfsInstance().Status.Id
	if fsId == "" {
		return nil, ctx
	}

	fs, err := state.client.DescribeFileSystem(ctx, fsId)
	if err != nil {
		logger.Error(err, "Error loading AliCloud NAS file system", "fileSystemId", fsId)
		return composed.StopWithRequeue, ctx
	}

	if fs != nil {
		state.fileSystem = fs
		state.fileSystemId = fs.FileSystemId
	}

	return nil, ctx
}
