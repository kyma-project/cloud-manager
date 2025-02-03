package gcpnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Removes finalizer from PVC when its parent GcpNfsVolume is marked for deletion
func removePersistenceVolumeClaimFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.IsMarkedForDeletion(state.PVC) {
		return nil, nil
	}

	if state.PVC == nil {
		return nil, nil
	}

	if !controllerutil.ContainsFinalizer(state.PVC, api.CommonFinalizerDeletionHook) {
		return nil, nil
	}

	controllerutil.RemoveFinalizer(state.PVC, api.CommonFinalizerDeletionHook)
	err := state.Cluster().K8sClient().Update(ctx, state.PVC)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR PersistentVolumeClaim after finalizer removal", composed.StopWithRequeue, ctx)
	}

	return nil, nil
}
