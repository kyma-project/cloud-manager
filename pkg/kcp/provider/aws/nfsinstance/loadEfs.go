package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/utils/pointer"
)

func loadEfs(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	list, err := state.awsClient.DescribeFileSystems(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error listing AWS file systems", composed.StopWithRequeue, nil)
	}

	for _, fs := range list {
		if pointer.StringDeref(fs.Name, "") == state.Obj().GetName() {
			state.efs = &fs
			break
		}
	}

	if state.efs == nil {
		return nil, nil
	}

	if len(state.ObjAsNfsInstance().Status.Id) > 0 {
		return nil, nil
	}

	state.ObjAsNfsInstance().Status.Id = *state.efs.FileSystemId
	err = state.UpdateObjStatus(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating NfsInstance status with file system id", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
