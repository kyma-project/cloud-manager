package sapnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func pvcRemoveFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}

	if state.PVC == nil {
		return nil, ctx
	}

	removed, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, state.PVC, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error removing finalizer from SapNfsVolume PVC", composed.StopWithRequeue, ctx)
	}
	if removed {
		logger.Info("Removed SapNfsVolume PVC finalizer")
	}

	return nil, ctx
}
