package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func removeKymaFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if state.kyma == nil {
		return nil, ctx
	}

	removed, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, state.kyma, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Kyma CR with removed finalizer", composed.StopWithRequeue, ctx)
	}

	if removed {
		logger.Info("Removed finalizer from the Kyma CR")
	}

	return nil, ctx
}
