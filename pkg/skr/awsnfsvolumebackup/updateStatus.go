package awsnfsvolumebackup

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func updateStatus(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	backup := state.ObjAsAwsNfsVolumeBackup()

	logger := composed.LoggerFromCtx(ctx)

	//Handle if deletion is in progress
	deleting := composed.IsMarkedForDeletion(backup)
	if deleting {
		//If the backup not already exists, return
		if state.recoveryPoint == nil {
			return composed.StopAndForget, nil
		} else {
			return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
		}
	}

	//Get Status from AWS
	var jobStatus types.BackupJobState
	var recoveryPointStatus types.RecoveryPointStatus
	if state.backupJob != nil {
		jobStatus = state.backupJob.State
	}
	if state.recoveryPoint != nil {
		recoveryPointStatus = state.recoveryPoint.Status
	}

	//If backup is completed, mark the backup as ready.
	if (jobStatus == "" || jobStatus == types.BackupJobStateCompleted) &&
		recoveryPointStatus == types.RecoveryPointStatusCompleted {

		//Check if the backup has a ready condition
		backupReady := meta.FindStatusCondition(backup.Status.Conditions, cloudresourcesv1beta1.ConditionTypeReady)
		if backupReady != nil && backupReady.Status == metav1.ConditionTrue &&
			backup.Status.State == cloudresourcesv1beta1.StateReady {
			// already with Ready condition
			return composed.StopAndForget, nil
		} else {
			logger.Info("Updating SKR AwsNfsVolumeBackup status with Ready condition")
			backup.Status.State = cloudresourcesv1beta1.StateReady
			return composed.UpdateStatus(backup).
				SetExclusiveConditions(metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeReady,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionTypeReady,
					Message: "Backup is ready for use.",
				}).
				SuccessLogMsg("AwsNfsVolumeBackup status got updated with Ready condition ").
				SuccessError(composed.StopAndForget).
				Run(ctx, state)
		}
	} else if jobStatus == types.BackupJobStateAborting ||
		jobStatus == types.BackupJobStateAborted ||
		jobStatus == types.BackupJobStatePartial ||
		jobStatus == types.BackupJobStateFailed ||
		jobStatus == types.BackupJobStateExpired ||
		recoveryPointStatus == types.RecoveryPointStatusExpired ||
		recoveryPointStatus == types.RecoveryPointStatusPartial {

		logger.Info("Updating SKR AwsNfsVolumeBackup status with Error condition")
		message := "AwsNfsVolumeBackup failed."
		if state.recoveryPoint != nil {
			message = fmt.Sprintf("AWS RecoveryPoint State: %s. Message: %s", state.recoveryPoint.Status, ptr.Deref(state.recoveryPoint.StatusMessage, ""))
		} else if state.backupJob != nil {
			message = fmt.Sprintf("AWS BackupJob State: %s. Message: %s", state.backupJob.State, ptr.Deref(state.backupJob.StatusMessage, ""))
		}
		backup.Status.State = cloudresourcesv1beta1.StateError
		return composed.UpdateStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeError,
				Message: message,
			}).
			SuccessLogMsg("AwsNfsVolumeBackup status got updated with error :"+message).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	} else {
		// Delay the requeue to check the status of the backup
		return composed.StopWithRequeueDelay(util.Timing.T1000ms()), nil
	}
}
