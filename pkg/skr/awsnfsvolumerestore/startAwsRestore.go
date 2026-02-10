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

	//Create a Restore Job
	restoreJobOutput, err := state.awsClient.StartRestoreJob(ctx, &client.StartRestoreJobInput{
		BackupVaultName:  state.GetVaultName(),
		IamRoleArn:       state.GetBackupRoleArn(),
		IdempotencyToken: ptr.To(restore.Status.IdempotencyToken),
		RecoveryPointArn: restoreMetadataOut.RecoveryPointArn,
		RestoreMetadata:  restoreMetadataOut.RestoreMetadata,
	})
	if err != nil {
		// If idempotency token was already used, the job likely exists but we lost the JobId
		// This can happen if status patch failed after StartRestoreJob succeeded
		// Skip with longer requeue delay to let checkRestoreJob handle it
		if strings.Contains(err.Error(), "Idempotency token already used") {
			logger.Info("Idempotency token already used - restore job may exist, will check status on next reconcile")
			return composed.StopWithRequeueDelay(util.Timing.T10000ms()), nil
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
