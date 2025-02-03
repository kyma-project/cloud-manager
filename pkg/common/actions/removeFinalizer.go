package actions

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func PatchRemoveCommonFinalizer() composed.Action {
	// Until the pkg/migrateFinalizers is executed there's a risk old finalizers still would be present
	// so to be sure we removed them all, have to strip all three of them, two old, and one new
	// The migration is executed concurrently with reconcilers, and it can add new finalizers on objects with deletion timestamp
	// so there's a risk some of the resources being deleted will remain with old finalizers
	return composed.ComposeActionsNoName(
		PatchRemoveFinalizer(api.DO_NOT_USE_OLD_KcpFinalizer),
		PatchRemoveFinalizer(api.DO_NOT_USE_OLD_SkrFinalizer),
		PatchRemoveFinalizer(api.CommonFinalizerDeletionHook),
	)
}

func PatchRemoveFinalizer(f string) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		if !composed.MarkedForDeletionPredicate(ctx, state) {
			return nil, ctx
		}

		removed, err := state.PatchObjRemoveFinalizer(ctx, f)
		if err != nil {
			return composed.LogErrorAndReturn(err, fmt.Sprintf("Error patching obj to remove finalizer: %s", f), composed.StopWithRequeue, ctx)
		}

		if removed {
			logger := composed.LoggerFromCtx(ctx)
			logger.Info(fmt.Sprintf("Finalizer %s patch removed", f))
		}

		return nil, ctx
	}
}

func RemoveCommonFinalizer() composed.Action {
	// Until the pkg/migrateFinalizers is executed there's a risk old finalizers still would be present
	// For details check PatchRemoveCommonFinalizer()
	return composed.ComposeActionsNoName(
		RemoveFinalizer(api.DO_NOT_USE_OLD_KcpFinalizer),
		RemoveFinalizer(api.DO_NOT_USE_OLD_SkrFinalizer),
		RemoveFinalizer(api.CommonFinalizerDeletionHook),
	)
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
		logger.Info(fmt.Sprintf("RemoveFinalizer: %s", f))

		if err := state.UpdateObj(ctx); err != nil {
			return composed.LogErrorAndReturn(err, fmt.Sprintf("Error removing Finalizer: %s", f), composed.StopWithRequeue, ctx)
		}

		return nil, ctx
	}
}
