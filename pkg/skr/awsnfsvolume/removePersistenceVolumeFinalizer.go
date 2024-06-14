package awsnfsvolume

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// Removes finalizer from PV when its parent AwsNfsVolume is marked for deletion
func removePersistenceVolumeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.IsMarkedForDeletion(state.Volume) {
		return nil, nil
	}

	if state.Volume == nil {
		return nil, nil
	}

	if !controllerutil.ContainsFinalizer(state.Volume, v1beta1.Finalizer) {
		return nil, nil
	}

	controllerutil.RemoveFinalizer(state.Volume, v1beta1.Finalizer)
	err := state.Cluster().K8sClient().Update(ctx, state.Volume)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR PersistentVolume after finalizer removal", composed.StopWithRequeue, ctx)
	}

	return nil, nil

}
