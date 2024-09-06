package gcpnfsvolume

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func addFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	modified, err := st.PatchObjAddFinalizer(ctx, cloudresourcesv1beta1.Finalizer)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving object after finalizer added", composed.StopWithRequeue, ctx)
	}
	if modified {
		composed.LoggerFromCtx(ctx).Info("Finalizer added")
	}

	return nil, nil
}
