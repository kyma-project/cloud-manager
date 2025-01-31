package gcpnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func addFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	modified, err := st.PatchObjAddFinalizer(ctx, api.CommonFinalizerDeletionHook)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving object after finalizer added", composed.StopWithRequeue, ctx)
	}
	if modified {
		composed.LoggerFromCtx(ctx).Info("Finalizer added")
	}

	return nil, nil
}
