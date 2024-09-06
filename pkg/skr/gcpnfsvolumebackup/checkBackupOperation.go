package gcpnfsvolumebackup

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/api/googleapi"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	cloudControl "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

func checkBackupOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	backup := state.ObjAsGcpNfsVolumeBackup()
	opName := backup.Status.OpIdentifier
	logger.WithValues("NfsBackup :", backup.Name).Info("Checking GCP Backup Operation Status")

	//If no OpIdentifier, then continue to next action.
	if opName == "" {
		return nil, nil
	}

	project := state.Scope.Spec.Scope.Gcp.Project
	op, err := state.fileBackupClient.GetBackupOperation(ctx, project, opName)
	if err != nil {

		//If the operation is not found, reset the OpIdentifier and let retry or updateStatus to update the status.
		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 404 {
				backup.Status.OpIdentifier = ""
			}
		}
		backup.Status.State = v1beta1.GcpNfsBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudControl.ReasonGcpError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Error getting Filestore backup Operation from GCP.").
			Run(ctx, state)
	}

	//Operation not completed yet.. requeue again.
	if op != nil && !op.Done {

		return composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), nil
	}

	//If not able to find the operation or it is completed, reset OpIdentifier.
	backup.Status.OpIdentifier = ""
	if op == nil {
		backup.Status.State = v1beta1.GcpNfsBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:   v1beta1.ConditionTypeError,
				Status: metav1.ConditionTrue,
				Reason: v1beta1.ConditionReasonError,
				Message: fmt.Sprintf("Could not find the backup operation %s in GCP to verify its state. "+
					"Retrying a new backup operation if backup with the same name is not found", opName),
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Error getting Filestore backup Operation from GCP.").
			Run(ctx, state)
	}

	//If the operation failed, update the error status on the object.
	if op != nil && op.Error != nil {
		backup.Status.State = v1beta1.GcpNfsBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudControl.ReasonGcpError,
				Message: op.Error.Message,
			}).
			SuccessError(composed.StopAndForget). //To reduce the rate. Next SKR looper with retry it.
			SuccessLogMsg(fmt.Sprintf("Filestore Operation error : %s", op.Error.Message)).
			Run(ctx, state)
	}

	//Done Successfully
	if backup.Status.State == v1beta1.GcpNfsBackupDeleting {
		backup.Status.State = v1beta1.GcpNfsBackupDeleted
		state.fileBackup = nil
		return composed.PatchStatus(backup).
			SetExclusiveConditions().
			SuccessErrorNil().
			Run(ctx, state)
	} else {
		// The only remaining operation is backup creation.
		backup.Status.State = v1beta1.GcpNfsBackupReady
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ConditionReasonReady,
				Message: fmt.Sprintf("Backup operation finished successfully: %s", opName),
			}).
			OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
				return nil, nil
			}). //proceed in case deletion is in progress
			SuccessLogMsg("GcpNfsVolumeBackup status got updated with Ready condition and Done state.").
			Run(ctx, state)
	}

}
