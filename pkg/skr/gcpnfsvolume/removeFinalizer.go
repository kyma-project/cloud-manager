package gcpnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	if state.KcpNfsInstance != nil {
		// KCP NfsInstance is not yet deleted
		return nil, nil
	}

	// KCP NfsInstance does not exist, remove the finalizer so SKR GcpNfsVolume is also deleted
	controllerutil.RemoveFinalizer(state.Obj(), api.CommonFinalizerDeletionHook)
	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR GcpNfsVolume after finalizer remove", composed.StopWithRequeue, ctx)
	}

	// bye, bye SKR GcpNfsVolume
	return composed.StopAndForget, nil
}
