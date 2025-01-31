package actions

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func PatchAddCommonFinalizer() composed.Action {
	return PatchAddFinalizer(api.CommonFinalizerDeletionHook)
}

func PatchAddFinalizer(f string) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		if composed.MarkedForDeletionPredicate(ctx, state) {
			return nil, ctx
		}

		_, err := state.PatchObjAddFinalizer(ctx, f)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error patch adding finalizer", composed.StopWithRequeue, ctx)
		}

		return nil, ctx
	}
}

func AddCommonFinalizer() composed.Action {
	return AddFinalizer(api.CommonFinalizerDeletionHook)
}

func AddFinalizer(f string) composed.Action {
	return func(ctx context.Context, state composed.State) (error, context.Context) {
		if composed.MarkedForDeletionPredicate(ctx, state) {
			return nil, nil
		}

		added := controllerutil.AddFinalizer(state.Obj(), f)
		if !added {
			return nil, ctx
		}

		if err := state.UpdateObj(ctx); err != nil {
			return composed.LogErrorAndReturn(err, "Error adding Finalizer", composed.StopWithRequeue, ctx)
		}

		return nil, ctx
	}
}
