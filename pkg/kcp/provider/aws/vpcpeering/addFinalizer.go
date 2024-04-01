package vpcpeering

import (
	"context"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func addFinalizer(ctx context.Context, state composed.State) (error, context.Context) {
	// Object is being deleted, don't add finalizer
	if composed.MarkedForDeletionPredicate(ctx, state) {
		return nil, nil
	}

	// If finalizer already present, don't add it again.
	if controllerutil.ContainsFinalizer(state.Obj(), cloudcontrolv1beta1.FinalizerName) {
		return nil, nil
	}

	//Add finalizer
	controllerutil.AddFinalizer(state.Obj(), cloudcontrolv1beta1.FinalizerName)

	if err := state.UpdateObj(ctx); err != nil {
		return composed.LogErrorAndReturn(err, "Error adding Finalizer", composed.StopWithRequeue, ctx)
	}

	// Requeue to reload the object
	return composed.StopWithRequeue, nil
}
