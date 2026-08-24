package alicloudnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	hasFinalizer := controllerutil.ContainsFinalizer(state.ObjAsAlicloudNfsVolume(), api.CommonFinalizerDeletionHook)
	if !hasFinalizer {
		return nil, nil
	}

	controllerutil.RemoveFinalizer(state.ObjAsAlicloudNfsVolume(), api.CommonFinalizerDeletionHook)
	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AlicloudNfsVolume after finalizer removed", composed.StopWithRequeue, ctx)
	}

	logger.Info("Finalizer removed")

	return composed.StopAndForget, nil
}
