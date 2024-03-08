package awsnfsvolume

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	if !composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	hasFinalizer := controllerutil.ContainsFinalizer(state.ObjAsAwsNfsVolume(), cloudresourcesv1beta1.Finalizer)
	if !hasFinalizer {
		return nil, nil
	}

	controllerutil.RemoveFinalizer(state.ObjAsAwsNfsVolume(), cloudresourcesv1beta1.Finalizer)
	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error updating SKR AwsNfsVolume after finalizer removed", composed.StopWithRequeue, ctx)
	}

	logger.Info("Finalizer removed")

	return composed.StopAndForget, nil
}
