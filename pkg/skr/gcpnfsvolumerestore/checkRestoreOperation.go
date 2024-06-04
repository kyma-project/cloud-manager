package gcpnfsvolumerestore

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

func checkRestoreOperation(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)

	restore := state.ObjAsGcpNfsVolumeRestore()
	opName := restore.Status.OpIdentifier
	logger.WithValues("Nfs Restore source:", restore.Spec.Source.Backup.ToNamespacedName(state.Obj().GetNamespace()),
		"destination:", restore.Spec.Destination.Volume.ToNamespacedName(state.Obj().GetNamespace())).Info("Checking GCP Restore Operation Status")

	//If no OpIdentifier, then continue to next action.
	if opName == "" {
		return nil, nil
	}

	project := state.Scope.Spec.Scope.Gcp.Project
	op, err := state.fileRestoreClient.GetRestoreOperation(ctx, project, opName)
	if err != nil {

		//If the operation is not found, reset the OpIdentifier.
		var e *googleapi.Error
		if ok := errors.As(err, &e); ok {
			if e.Code == 404 {
				restore.Status.OpIdentifier = ""
			}
		}
		restore.Status.State = v1beta1.JobStateError
		return composed.UpdateStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ConditionReasonNfsRestoreFailed,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Error getting Filestore restore Operation from GCP.").
			Run(ctx, state)
	}

	//Operation not completed yet.. requeue again.
	if op != nil && !op.Done {

		return composed.StopWithRequeueDelay(state.gcpConfig.GcpRetryWaitTime), nil
	}

	//If not able to find the operation or it is completed, reset OpIdentifier.
	restore.Status.OpIdentifier = ""
	if op == nil {
		restore.Status.State = v1beta1.JobStateError
		return composed.UpdateStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  v1beta1.ConditionReasonNfsRestoreFailed,
				Message: fmt.Sprintf("Could not find the restore operation %s in GCP to verify its state. Retrying a new restore operation", opName),
			}).
			SuccessError(composed.StopWithRequeue).
			SuccessLogMsg("Error getting Filestore restore Operation from GCP.").
			Run(ctx, state)
	}

	//If the operation failed, update the error status on the object.
	if op != nil && op.Error != nil {
		restore.Status.State = v1beta1.JobStateFailed
		return composed.UpdateStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    v1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudControl.ReasonGcpError,
				Message: op.Error.Message,
			}).
			OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
				return nil, nil
			}). //proceed in case deletion is in progress
			SuccessLogMsg(fmt.Sprintf("Filestore Operation error : %s", op.Error.Message)).
			Run(ctx, state)
	}

	//Done Successfully
	restore.Status.State = v1beta1.JobStateDone
	return composed.UpdateStatus(restore).
		SetExclusiveConditions(metav1.Condition{
			Type:    v1beta1.ConditionTypeReady,
			Status:  metav1.ConditionTrue,
			Reason:  v1beta1.ConditionReasonReady,
			Message: fmt.Sprintf("Restore operation finished successfully: %s", opName),
		}).
		OnUpdateSuccess(func(ctx context.Context) (error, context.Context) {
			return nil, nil
		}). //proceed in case deletion is in progress
		SuccessLogMsg("GcpNfsVolumeRestore status got updated with Ready condition and Done state.").
		Run(ctx, state)
}
