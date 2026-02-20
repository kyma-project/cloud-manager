package v1

import (
	"context"

	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func shortCircuitCompleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)

	//If deletion, continue.
	if composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, nil
	}

	backup := state.ObjAsGcpNfsVolumeBackup()
	backupState := backup.Status.State
	if backupState == v1beta1.GcpNfsBackupFailed {
		composed.LoggerFromCtx(ctx).Info("NfsVolumeBackup is failed , short-circuiting into StopAndForget")
		return composed.StopAndForget, nil
	}

	if state.ShouldShortCircuit() {
		composed.LoggerFromCtx(ctx).Info("NfsVolumeBackup is ready and accessibleFrom hasn't changed, short-circuiting into requeueForCapacity")
		return stopAndRequeueForCapacity(), nil
	}

	return nil, ctx
}
