package awsnfsvolumebackup

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadAwsBackup(ctx context.Context, st composed.State, local bool) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()
	deleting := composed.IsMarkedForDeletion(backup)

	id := backup.Status.Id
	client := state.awsClient
	arn := state.GetRecoveryPointArn()
	if !local {
		id = backup.Status.RemoteId
		client = state.destAwsClient
		arn = state.GetDestinationRecoveryPointArn()
	}

	//Backup Id is empty, Continue.
	if len(id) == 0 {
		return nil, ctx
	}

	// Load the backup from AWS
	logger.WithValues("local", local).Info("Loading AWS RecoveryPoint")

	//Load the recovery point from AWS
	recoveryPoint, err := client.DescribeRecoveryPoint(ctx,
		state.Scope().Spec.Scope.Aws.AccountId,
		state.GetVaultName(),
		arn)

	if err == nil {
		//store the recoveryPoint in the state object
		if local {
			state.recoveryPoint = recoveryPoint
		} else {
			state.destRecoveryPoint = recoveryPoint
		}
		return nil, ctx
	}

	// If deleting and not found, continue...
	if deleting && client.IsNotFound(err) {
		return nil, ctx
	}

	//Update the status with error, and stop reconciliation
	backup.Status.State = cloudresourcesv1beta1.StateError
	return composed.PatchStatus(backup).
		SetExclusiveConditions(metav1.Condition{
			Type:    cloudresourcesv1beta1.ConditionTypeError,
			Status:  metav1.ConditionTrue,
			Reason:  cloudcontrolv1beta1.ConditionTypeError,
			Message: err.Error(),
		}).
		SuccessLogMsg(fmt.Sprintf("Error loading the Recovery Point : %s", err)).
		SuccessError(composed.StopAndForget).
		Run(ctx, state)
}

func loadLocalAwsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	return loadAwsBackup(ctx, st, true)
}

func loadDestAwsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	return loadAwsBackup(ctx, st, false)
}
