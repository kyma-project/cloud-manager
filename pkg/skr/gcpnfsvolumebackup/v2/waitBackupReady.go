package v2

import (
	"context"

	"cloud.google.com/go/filestore/apiv1/filestorepb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/config"
)

// waitBackupReady waits for the backup to be in READY state.
// This is a redundancy action in case OpIdentifier wasn't properly saved.
func waitBackupReady(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	// If the backup doesn't exist yet, let createNfsBackup handle it
	if state.fileBackup == nil {
		return nil, nil
	}

	// Check if backup is ready
	if state.fileBackup.State == filestorepb.Backup_READY {
		logger.Info("Backup is in READY state")
		return nil, nil
	}

	// If backup is still creating, wait
	if state.fileBackup.State == filestorepb.Backup_CREATING {
		logger.Info("Backup is still CREATING, requeuing...")
		return composed.StopWithRequeueDelay(config.GcpConfig.GcpRetryWaitTime), nil
	}

	// For any other state, continue and let updateStatus handle it
	logger.Info("Backup state", "state", state.fileBackup.State.String())
	return nil, nil
}
