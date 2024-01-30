package gcpnfsvolume

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removePersistenceVolumeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//If PV doesn't exist OR is not marked for deletion, continue
	if state.PV == nil || state.PV.DeletionTimestamp.IsZero() {
		return nil, nil
	}

	//If Kcp NfsInstance still exists, continue
	if state.KcpNfsInstance != nil {
		// KCP NfsInstance is not yet deleted
		return nil, nil
	}

	// KCP NfsInstance does not exist, remove the finalizer so PV is also deleted
	controllerutil.RemoveFinalizer(state.PV, cloudresourcesv1beta1.Finalizer)
	err := state.SkrCluster.K8sClient().Update(ctx, state.PV)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR PersistentVolume after finalizer removal", composed.StopWithRequeue, nil)
	}

	// bye, bye SKR PersistentVolume
	return nil, nil

}
