package awsnfsvolumebackup

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func deleteAwsBackup(ctx context.Context, st composed.State, local bool) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()

	client := state.awsClient
	arn := state.GetRecoveryPointArn()
	recoveryPoint := state.recoveryPoint
	if !local {
		recoveryPoint = state.destRecoveryPoint
		client = state.destAwsClient
		arn = state.GetDestinationRecoveryPointArn()
	}

	//If not deleting, return.
	if !composed.IsMarkedForDeletion(backup) {
		return nil, ctx
	}

	//If the backup not already exists, return
	if recoveryPoint == nil {
		return nil, ctx
	}

	logger.WithValues("AwsBackup", backup.Name, "local", local).Info("Deleting AWS File RecoveryPoint")

	_, err := client.DeleteRecoveryPoint(ctx, state.GetVaultName(), arn)

	//If failed, update status with error state.
	if err != nil && !client.IsNotFound(err) {
		backup.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: err.Error(),
			}).
			SuccessError(err).
			SuccessLogMsg(fmt.Sprintf("Error deleting RecoveryPoint object in AWS :%s", err)).
			Run(ctx, state)
	}

	//update status with deleting state.
	backup.Status.State = cloudresourcesv1beta1.StateDeleting
	if !local {
		backup.Status.State = cloudresourcesv1beta1.StateDeletingRemote
	}
	return composed.PatchStatus(backup).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}

func deleteLocalAwsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	return deleteAwsBackup(ctx, st, true)
}

func deleteDestAwsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	return deleteAwsBackup(ctx, st, false)
}
