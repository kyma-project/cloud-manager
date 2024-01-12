package focal

import (
	"context"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func AddFinalizer(ctx context.Context, state composed.State) (error, context.Context) {

	//Object is being deleted, don't add finalizer
	if state.Obj().GetDeletionTimestamp().IsZero() {
		return nil, nil
	}

	//If finalizer already present, don't add it again.
	if controllerutil.ContainsFinalizer(state.Obj(), cloudresourcesv1beta1.FinalizerName) {
		return nil, nil
	}

	//Add finalizer
	controllerutil.AddFinalizer(state.Obj(), cloudresourcesv1beta1.FinalizerName)
	if err := state.UpdateObj(ctx); err != nil {
		return composed.LogErrorAndReturn(err, "Error adding Finalizer", composed.StopWithRequeue, nil)
	}

	//continue
	return nil, nil
}
