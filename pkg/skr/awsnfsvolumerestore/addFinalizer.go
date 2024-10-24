package awsnfsvolumerestore

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func addFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	added := controllerutil.AddFinalizer(st.Obj(), cloudresourcesv1beta1.Finalizer)
	if !added {
		// finalizer already added
		return nil, nil
	}

	err := st.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving object after finalizer added", composed.StopWithRequeue, ctx)
	}

	return composed.StopWithRequeue, nil
}
