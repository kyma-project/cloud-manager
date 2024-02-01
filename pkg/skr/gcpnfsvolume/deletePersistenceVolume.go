package gcpnfsvolume

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"time"
)

func deletePersistenceVolume(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)

	//If GcpNfsVolume is not marked for deletion, continue
	if !composed.MarkedForDeletionPredicate(ctx, st) {
		// SKR GcpNfsVolume is NOT marked for deletion, do not delete mirror in KCP
		return nil, nil
	}

	//If PV doesn't exist or already marked for Deletion, continue
	if state.PV == nil || !state.PV.DeletionTimestamp.IsZero() {
		return nil, nil
	}

	//Delete PV
	err := state.SkrCluster.K8sClient().Delete(ctx, state.PV)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error deleting PersistentVolume", composed.StopWithRequeue, nil)
	}

	// give some time, and then run again
	return composed.StopWithRequeueDelay(3 * time.Second), nil
}
