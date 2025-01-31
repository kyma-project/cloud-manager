package actions

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func PatchRemoveCommonFinalizer() composed.Action {
	return PatchRemoveFinalizer(api.CommonFinalizerDeletionHook)
}

func PatchRemoveFinalizer(f string) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		if !composed.MarkedForDeletionPredicate(ctx, state) {
			return nil, ctx
		}

		removed, err := state.PatchObjRemoveFinalizer(ctx, f)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error patching obj to remove finalizer", composed.StopWithRequeue, ctx)
		}

		if removed {
			logger := composed.LoggerFromCtx(ctx)
			logger.Info("Finalizer patch removed")
		}

		return nil, ctx
	}
}

func RemoveCommonFinalizer() composed.Action {
	return RemoveFinalizer(api.CommonFinalizerDeletionHook)
}

func RemoveFinalizer(f string) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		if !composed.MarkedForDeletionPredicate(ctx, state) {
			return nil, nil
		}

		removed := controllerutil.RemoveFinalizer(state.Obj(), f)
		if !removed {
			return nil, ctx
		}

		logger := composed.LoggerFromCtx(ctx)
		logger.Info("RemoveFinalizer")

		if err := state.UpdateObj(ctx); err != nil {
			return composed.LogErrorAndReturn(err, "Error removing Finalizer", composed.StopWithRequeue, ctx)
		}

		return nil, ctx
	}
}
