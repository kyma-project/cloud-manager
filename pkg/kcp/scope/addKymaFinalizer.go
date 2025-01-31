package scope

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func addKymaFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !state.ObjAsScope().DeletionTimestamp.IsZero() {
		// Scope is being deleted
		return nil, ctx
	}

	if !state.kyma.GetDeletionTimestamp().IsZero() {
		// kyma is being deleted - it has deletionTimestamp and finalizer can not be added in that state
		return nil, ctx
	}

	added, err := composed.PatchObjAddFinalizer(ctx, api.CommonFinalizerDeletionHook, state.kyma, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating Kyma CR with added finalizer", composed.StopWithRequeue, ctx)
	}

	if added {
		logger.Info("Added finalizer to the Kyma CR")
	}

	return nil, ctx
}
