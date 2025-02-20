package awsnfsvolumebackup

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

	backup := state.ObjAsAwsNfsVolumeBackup()
	backupState := backup.Status.State
	if backupState == v1beta1.StateReady || backupState == v1beta1.StateFailed {
		composed.LoggerFromCtx(ctx).Info("NfsVolumeBackup is complete , short-circuiting into StopAndForget")
		return composed.StopAndForget, nil
	}

	return nil, ctx
}
