package awsnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Removes finalizer from PV when its parent AwsNfsVolume is marked for deletion
func removePersistenceVolumeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.IsMarkedForDeletion(state.Volume) {
		return nil, ctx
	}

	if state.Volume == nil {
		return nil, ctx
	}

	if !controllerutil.ContainsFinalizer(state.Volume, api.CommonFinalizerDeletionHook) {
		return nil, ctx
	}

	controllerutil.RemoveFinalizer(state.Volume, api.CommonFinalizerDeletionHook)
	err := state.Cluster().K8sClient().Update(ctx, state.Volume)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR PersistentVolume after finalizer removal", composed.StopWithRequeue, ctx)
	}

	return nil, ctx

}
