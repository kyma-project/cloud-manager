package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func scopeDelete(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.IsObjLoaded(ctx, state) {
		return nil, ctx
	}

	logger.Info("Deleting Scope")

	_, err := state.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook)
	if apierrors.IsNotFound(err) {
		return nil, ctx
	}
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error updating Scope after finalizer removed", composed.StopWithRequeue, ctx)
	}

	err = state.Cluster().K8sClient().Delete(ctx, state.Obj())
	if apierrors.IsNotFound(err) {
		return nil, ctx
	}
	if client.IgnoreNotFound(err) != nil {
		return composed.LogErrorAndReturn(err, "Error deleting Scope", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
