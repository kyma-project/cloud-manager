package azurerwxvolumerestore

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func checkRestoreJob(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsAzureRwxVolumeRestore()
	logger := composed.LoggerFromCtx(ctx)
	if restore.Status.OpIdentifier == "" {
		return composed.LogErrorAndReturn(nil, "Should not reach checkRestoreJob action if opIdentifier is missing.", composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx)
	}
	logger.Info("Checking restore job status", "opIdentifier", restore.Status.OpIdentifier)
	_, resourceGroup, vault, _, _, _, err := client.ParseRecoveryPointId(state.azureRwxVolumeBackup.Status.RecoveryPointId)
	if err != nil {
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonInvalidRecoveryPointId,
				Message: fmt.Sprintf("Source AzureRwxVolumeBackup has an invalid recoveryPointId: '%v'", state.azureRwxVolumeBackup.Status.RecoveryPointId),
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	job, err := state.storageClient.GetStorageJob(ctx, vault, resourceGroup, restore.Status.OpIdentifier)

	if err != nil && !meta.IsNotFound(err) {
		return composed.LogErrorAndReturn(err, "Error getting restore job", composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx)
	}
	if err != nil || job == nil {
		logger.Error(nil, "Restore job not found. Remove OpIdentifier and retry ")
		restore.Status.OpIdentifier = ""
		return composed.PatchStatus(restore).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}
	if job.Status == nil {
		return composed.LogErrorAndReturn(nil, "Restore job status is nil. Retry later", composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx)
	}
	switch *job.Status {
	case string(armrecoveryservicesbackup.JobStatusInProgress):
		logger.Info("Restore job in progress")
		if restore.Status.State == cloudresourcesv1beta1.JobStateProcessing {
			restore.Status.State = cloudresourcesv1beta1.JobStateInProgress
			return composed.PatchStatus(restore).SuccessError(composed.StopWithRequeueDelay(util.Timing.T10000ms())).Run(ctx, state)
		}
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	case string(armrecoveryservicesbackup.JobStatusFailed):
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		message, err := client.AzureStorageErrorInfoToJson(job.ErrorDetails)
		if err != nil {
			logger.Error(err, "Error in marshalling restore job error details to json")
			message = "Could not get error details"
		}
		logger.Error(nil, "Restore job failed", "message", message)
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonRestoreJobFailed,
				Message: fmt.Sprintf("Restore operation failed: %v", message),
			}).
			Run(ctx, state)
	case string(armrecoveryservicesbackup.JobStatusCancelled):
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		message, err := client.AzureStorageErrorInfoToJson(job.ErrorDetails)
		if err != nil {
			logger.Error(err, "Error in marshalling restore job error details to json")
			message = "Could not get error details"
		}
		logger.Error(nil, "Restore job cancelled", "message", message)
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonRestoreJobCancelled,
				Message: fmt.Sprintf("Restore operation got cancelled: %v", message),
			}).
			Run(ctx, state)

	case string(armrecoveryservicesbackup.JobStatusCancelling):
		logger.Info("Restore job is in 'cancelling' state. Wait to reach final status.")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	case string(armrecoveryservicesbackup.JobStatusCompleted):
		restore.Status.State = cloudresourcesv1beta1.JobStateDone
		logger.Info("Restore job completed")
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonReady,
				Message: "Restore operation finished successfully.",
			}).
			Run(ctx, state)
	case string(armrecoveryservicesbackup.JobStatusCompletedWithWarnings):
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		message, err := client.AzureStorageErrorInfoToJson(job.ErrorDetails)
		if err != nil {
			logger.Error(err, "Error in marshalling restore job error details to json")
			message = "Could not get warning details"
		}
		logger.Error(nil, "Restore job completed with warnings", "message", message)
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonRestoreJobCompletedWithWarnings,
				Message: fmt.Sprintf("The correctness of restore is not guarantied. Restore operation completed with warnings: %v", message),
			}).
			Run(ctx, state)
	default:
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		message, err := client.AzureStorageErrorInfoToJson(job.ErrorDetails)
		if err != nil {
			logger.Error(err, "Error in marshalling restore job error details to json")
			message = "Could not get error details"
		}
		logger.Error(nil, "Restore job is in unexpected status.", "message", message)
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonRestoreJobInvalidStatus,
				Message: fmt.Sprintf("Restore operation is in unexpected status: %v", message),
			}).
			Run(ctx, state)
	}
}
