package awsnfsvolumebackup

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func loadAwsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()
	deleting := composed.IsMarkedForDeletion(backup)

	//Backup Id is empty, Continue.
	if backup.Status.Id == "" {
		return nil, nil
	}

	// Load the backup from AWS
	logger.Info("Loading AWS Backup")
	backupJob, err := state.awsClient.DescribeBackupJob(ctx, backup.Status.JobId)
	if err != nil && !state.awsClient.IsNotFound(err) {
		return composed.LogErrorAndReturn(err, "Error loading AWS Backup Job", composed.StopWithRequeueDelay(util.Timing.T1000ms()), ctx)
	}

	//store the backupJob in the state object
	state.backupJob = backupJob

	//Load the recovery point from AWS
	recoveryPoint, err := state.awsClient.DescribeRecoveryPoint(ctx,
		state.Scope().Spec.Scope.Aws.AccountId,
		state.GetVaultName(),
		state.GetRecoveryPointArn())

	if err == nil {
		//store the recoveryPoint in the state object
		state.recoveryPoint = recoveryPoint
		return nil, nil
	}

	// If deleting and not found, continue...
	if deleting && state.awsClient.IsNotFound(err) {
		return nil, nil
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
