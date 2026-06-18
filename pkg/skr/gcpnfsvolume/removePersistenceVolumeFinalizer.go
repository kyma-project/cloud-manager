package gcpnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removePersistenceVolumeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//If PV doesn't exist OR PV is not marked for deletion, continue
	if state.PV == nil || !composed.IsMarkedForDeletion(state.PV) {
		return nil, ctx
	}

	//If finalizer not already present, don't remove it .
	if !controllerutil.ContainsFinalizer(state.PV, api.CommonFinalizerDeletionHook) {
		return nil, ctx
	}

	// KCP NfsInstance does not exist, remove the finalizer so PV is also deleted
	_, err := composed.PatchObjRemoveFinalizer(ctx, api.CommonFinalizerDeletionHook, state.PV, state.SkrCluster.K8sClient())
	if err != nil {
		if apierrors.IsConflict(err) {
			return composed.StopWithRequeue, nil
		}
		return composed.LogErrorAndReturn(err, "Error saving SKR PersistentVolume after finalizer removal", composed.StopWithRequeue, ctx)
	}

	// bye, bye SKR PersistentVolume
	return nil, ctx

}
