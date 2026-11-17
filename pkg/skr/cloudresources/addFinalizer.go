package cloudresources

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func addFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	added, err := composed.PatchObjAddFinalizer(ctx, api.CommonFinalizerDeletionHook, state.ObjAsCloudResources(), state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving object after finalizer added", composed.StopWithRequeue, ctx)
	}
	if added {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Added finalizer successfully")
	}

	return nil, ctx
}
