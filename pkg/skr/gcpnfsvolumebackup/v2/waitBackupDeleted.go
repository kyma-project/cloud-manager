package v2

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
)

// waitBackupDeleted waits for the backup to be deleted.
// This is a redundancy action in case OpIdentifier wasn't properly saved.
func waitBackupDeleted(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// If the backup doesn't exist, deletion is complete
	if state.fileBackup == nil {
		logger.Info("Backup is deleted (not found)")
		return nil, nil
	}

	// If backup is in DELETING state, wait
	if state.fileBackup.State == filestorepb.Backup_DELETING {
		logger.Info("Backup is still DELETING, requeuing...")
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
	}

	// Backup still exists but not in DELETING state - requeue to retry delete
	logger.Info("Backup exists but not in DELETING state, requeuing to retry delete", "state", state.fileBackup.State.String())
	return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
}
