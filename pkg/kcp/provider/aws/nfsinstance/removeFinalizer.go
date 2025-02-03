package nfsinstance

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	isUpdated := controllerutil.RemoveFinalizer(state.ObjAsNfsInstance(), api.CommonFinalizerDeletionHook)
	if !isUpdated {
		return nil, nil
	}

	logger.Info("Removing finalizer")

	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating KCP NfsInstance after finalizer removed", composed.StopWithRequeue, ctx)
	}

	return composed.StopAndForget, nil
}
