package awsnfsvolumebackup

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"time"
)

func createAwsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()

	//If the object is being deleted, continue...
	if composed.IsMarkedForDeletion(backup) {
		return nil, nil
	}

	//If the aws backup was already created, continue...
	if backup.Status.Id != "" || backup.Status.JobId != "" {
		return nil, nil
	}

	// Backup does not exist
	logger.Info("Creating AWS Backup")

	//Create a Backup Job
	res, err := state.awsClient.StartBackupJob(ctx, &client.StartBackupJobInput{
		BackupVaultName:   state.GetVaultName(),
		IamRoleArn:        state.roleName,
		ResourceArn:       state.GetFileSystemArn(),
		RecoveryPointTags: state.GetTags(),
	})
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error creating AWS Backup ", composed.StopWithRequeueDelay(time.Second), ctx)
	}

	//Update the status with details.
	backup.Status.JobId = ptr.Deref(res.BackupJobId, "")
	backup.Status.Id = state.awsClient.ParseRecoveryPointId(ptr.Deref(res.RecoveryPointArn, ""))
	backup.Status.State = cloudresourcesv1beta1.StateCreating
	return composed.UpdateStatus(backup).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
		Run(ctx, state)
}
