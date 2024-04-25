package gcpnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removePersistenceVolumeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//If PV doesn't exist OR PV is not marked for deletion, continue
	if state.PV == nil || !composed.IsMarkedForDeletion(state.PV) {
		return nil, nil
	}

	//If finalizer not already present, don't remove it .
	if !controllerutil.ContainsFinalizer(state.PV, v1beta1.Finalizer) {
		return nil, nil
	}

	// KCP NfsInstance does not exist, remove the finalizer so PV is also deleted
	controllerutil.RemoveFinalizer(state.PV, v1beta1.Finalizer)
	err := state.SkrCluster.K8sClient().Update(ctx, state.PV)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR PersistentVolume after finalizer removal", composed.StopWithRequeue, ctx)
	}

	// bye, bye SKR PersistentVolume
	return nil, nil

}
