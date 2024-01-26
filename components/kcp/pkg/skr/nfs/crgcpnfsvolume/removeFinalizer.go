package crgcpnfsvolume

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/components/lib/composed"
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
	controllerutil.RemoveFinalizer(state.Obj(), cloudresourcesv1beta1.Finalizer)
	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR GcpNfsVolume after finalizer remove", composed.StopWithRequeue, nil)
	}

	// bye, bye SKR GcpNfsVolume
	return composed.StopAndForget, nil
}
