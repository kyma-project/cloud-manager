package gcpnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Removes finalizer from PVC when its parent GcpNfsVolume is marked for deletion
func removePersistenceVolumeClaimFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.IsMarkedForDeletion(state.PVC) {
		return nil, ctx
	}

	if state.PVC == nil {
		return nil, ctx
	}

	if !controllerutil.ContainsFinalizer(state.PVC, api.CommonFinalizerDeletionHook) {
		return nil, ctx
	}

	_, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, state.PVC, state.Cluster().K8sClient())
	if err != nil {
		if apierrors.IsConflict(err) {
			return composed.StopWithRequeue, nil
		}
		return composed.LogErrorAndReturn(err, "Error saving SKR PersistentVolumeClaim after finalizer removal", composed.StopWithRequeue, ctx)
	}

	return nil, ctx
}
