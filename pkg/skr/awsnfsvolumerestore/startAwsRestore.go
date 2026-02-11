package awsnfsvolumerestore

import (
	"context"
	"fmt"
	"strings"
	"time"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumerestore/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func startAwsRestore(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	restore := state.ObjAsAwsNfsVolumeRestore()

	//If the object is being deleted, continue...
	if composed.IsMarkedForDeletion(restore) {
		return nil, nil
	}

	//If the aws restore job already exists, continue...
	if restore.Status.JobId != "" {
		return nil, nil
	}

	// Restore does not exist
	logger.Info("Starting AWS Restore Job")

	//Get
	restoreMetadataOut, err := state.awsClient.GetRecoveryPointRestoreMetadata(ctx,
		state.Scope().Spec.Scope.Aws.AccountId,
		state.GetVaultName(),
		state.GetRecoveryPointArn())

	if err != nil {
		//Update the status with error, and stop reconciliation
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: err.Error(),
			}).
			SuccessLogMsg(fmt.Sprintf("Error loading the Recovery Point Restore Metadata: %s", err)).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	// Log the metadata we received from AWS
	logger.Info("Restore metadata from AWS", "metadata", restoreMetadataOut.RestoreMetadata)

	// AWS returns metadata that may include parameters for creating a NEW filesystem
	// We need to restore to the EXISTING filesystem, so set both file-system-id and newFileSystem
	// References: https://docs.aws.amazon.com/aws-backup/latest/devguide/restoring-efs.html
	fileSystemId := state.GetFileSystemId()
	if fileSystemId == "" {
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: "Cannot determine EFS filesystem ID from NFS volume",
			}).
			SuccessLogMsg("Missing filesystem ID").
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	// Set metadata to restore to existing filesystem
	// newFileSystem=false tells AWS to restore to existing EFS
	// AWS Backup will restore to a timestamped subdirectory by default
	restoreMetadataOut.RestoreMetadata = map[string]string{
		"file-system-id": fileSystemId,
		"newFileSystem":  "false",
	}

	logger.Info("Modified restore metadata for existing filesystem", "metadata", restoreMetadataOut.RestoreMetadata, "fileSystemId", fileSystemId)

	//Create a Restore Job
	restoreJobOutput, err := state.awsClient.StartRestoreJob(ctx, &client.StartRestoreJobInput{
		BackupVaultName:  state.GetVaultName(),
		IamRoleArn:       state.GetBackupRoleArn(),
		IdempotencyToken: ptr.To(restore.Status.IdempotencyToken),
		RecoveryPointArn: restoreMetadataOut.RecoveryPointArn,
		RestoreMetadata:  restoreMetadataOut.RestoreMetadata,
	})
	if err != nil {
		// If idempotency token was already used, the restore job was created but we lost the JobId
		// (e.g., pod crashed after StartRestoreJob but before status patch succeeded)
		// We cannot recover the JobId without ListRestoreJobs permission, so we'll fail
		// and require manual cleanup or CR recreation
		if strings.Contains(err.Error(), "Idempotency token already used") {
			restore.Status.State = cloudresourcesv1beta1.JobStateError
			return composed.PatchStatus(restore).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudcontrolv1beta1.ConditionTypeError,
					Message: "Idempotency token collision detected - restore job exists in AWS but JobId was lost. Please delete this CR and create a new one.",
				}).
				SuccessLogMsg("Idempotency token collision - cannot recover JobId").
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}
		return composed.LogErrorAndReturn(err, "Error starting AWS Restore", composed.StopWithRequeueDelay(time.Second), ctx)
	}
	//Update the status with details.
	restore.Status.JobId = ptr.Deref(restoreJobOutput.RestoreJobId, "")
	restore.Status.State = cloudresourcesv1beta1.JobStateInProgress
	return composed.PatchStatus(restore).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
		Run(ctx, state)
}
