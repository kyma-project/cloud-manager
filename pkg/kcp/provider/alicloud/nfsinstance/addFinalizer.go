package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func addFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	// Object is being deleted, don't add finalizer
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	if controllerutil.ContainsFinalizer(st.Obj(), api.CommonFinalizerDeletionHook) {
		return nil, ctx
	}

	controllerutil.AddFinalizer(st.Obj(), api.CommonFinalizerDeletionHook)
	if err := st.UpdateObj(ctx); err != nil {
		return composed.LogErrorAndReturn(err, "Error adding finalizer to AliCloud KCP NfsInstance", composed.StopWithRequeue, ctx)
	}

	// Requeue to reload the object
	return composed.StopWithRequeue, ctx
}
