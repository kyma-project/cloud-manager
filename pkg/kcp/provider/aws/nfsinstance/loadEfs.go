package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/meta"
	"k8s.io/utils/ptr"
)

func loadEfs(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	list, err := state.awsClient.DescribeFileSystems(ctx)
	if err != nil {
		return awsmeta.LogErrorAndReturn(err, "Error listing AWS file systems", ctx)
	}

	for _, fs := range list {
		if ptr.Deref(fs.Name, "") == state.Obj().GetName() {
			state.efs = &fs
			break
		}
	}

	if state.efs == nil {
		return nil, nil
	}

	logger = logger.WithValues(
		"efsId", ptr.Deref(state.efs.FileSystemId, ""),
		"efsLifeCycleState", state.efs.LifeCycleState,
	)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	if len(state.ObjAsNfsInstance().Status.Id) > 0 {
		return nil, ctx
	}

	state.ObjAsNfsInstance().Status.Id = *state.efs.FileSystemId
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating NfsInstance status with file system id", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, ctx
}
