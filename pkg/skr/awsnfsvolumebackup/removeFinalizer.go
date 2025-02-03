package awsnfsvolumebackup

import (
	"context"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func removeFinalizer(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	// If the backup still exists, continue...
	if state.recoveryPoint != nil {
		// GCP backup is not yet deleted
		return nil, nil
	}

	// AWS file backup does not exist, remove the finalizer so SKR GcpNfsVolumeBackup is also deleted
	controllerutil.RemoveFinalizer(state.Obj(), api.CommonFinalizerDeletionHook)
	err := state.UpdateObj(ctx)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error saving SKR AwsNfsVolumeBackup after finalizer remove", composed.StopWithRequeue, ctx)
	}

	// bye, bye SKR AwsNfsVolumeBackup
	return composed.StopAndForget, nil
}
