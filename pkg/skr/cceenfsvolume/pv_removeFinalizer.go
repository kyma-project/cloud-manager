package cceenfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func pvRemoveFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.IsMarkedForDeletion(state.Obj()) {
		return nil, ctx
	}

	if state.PV == nil {
		return nil, ctx
	}

	removed, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, state.PV, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error removing finalizer from CceeNfsVolume PV", composed.StopWithRequeue, ctx)
	}
	if removed {
		logger.Info("Removed CceeNfsVolume PV finalizer")
	}

	return nil, ctx
}
