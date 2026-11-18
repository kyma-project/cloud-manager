package awsnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"

	"github.com/kyma-project/cloud-manager/pkg/composed"
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

	_, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, state.Volume, state.Cluster().K8sClient())
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR PersistentVolume after finalizer removal", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
