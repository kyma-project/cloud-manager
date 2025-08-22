package awsnfsvolumebackup

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func loadAwsBackupJob(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()

	//If the object is being deleted continue...
	if composed.IsMarkedForDeletion(backup) {
		return nil, ctx
	}

	//If the state.backupJob is not empty continue
	if state.backupJob != nil {
		return nil, ctx
	}

	//If Status.JobId is empty, Continue.
	if len(backup.Status.JobId) == 0 {
		return nil, ctx
	}

	// Load the copyJob from AWS
	logger.Info("Loading AWS Backup Job")
	backupJob, err := state.awsClient.DescribeBackupJob(ctx, backup.Status.JobId)
	if err != nil && !state.awsClient.IsNotFound(err) {
		return composed.LogErrorAndReturn(err, "Error loading AWS Backup Job", err, ctx)
	}

	//store the backupJob in the state object
	state.backupJob = backupJob
	return nil, ctx
}
