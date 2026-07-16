package nfsinstance

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// removeFinalizer removes the common finalizer once all NAS resources are deleted.
func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	if !controllerutil.ContainsFinalizer(st.Obj(), api.CommonFinalizerDeletionHook) {
		return nil, ctx
	}

	controllerutil.RemoveFinalizer(st.Obj(), api.CommonFinalizerDeletionHook)
	if err := st.UpdateObj(ctx); err != nil {
		return composed.LogErrorAndReturn(err, "Error removing finalizer from AliCloud KCP NfsInstance", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
