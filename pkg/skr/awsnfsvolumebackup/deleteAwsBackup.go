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

func deleteAwsBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()

	//If not deleting, return.
	if !composed.IsMarkedForDeletion(backup) {
		return nil, nil
	}

	//If the backup not already exists, return
	if state.recoveryPoint == nil {
		return nil, nil
	}

	logger.WithValues("AwsBackup", backup.Name).Info("Deleting AWS File Backup")

	_, err := state.awsClient.DeleteRecoveryPoint(ctx, state.GetVaultName(), state.GetRecoveryPointArn())

	//If failed, update status with error state.
	if err != nil && !state.awsClient.IsNotFound(err) {
		backup.Status.State = cloudresourcesv1beta1.StateError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeError,
				Message: err.Error(),
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
			SuccessLogMsg(fmt.Sprintf("Error deleting RecoveryPoint object in AWS :%s", err)).
			Run(ctx, state)
	}

	//update status with deleting state.
	backup.Status.State = cloudresourcesv1beta1.StateDeleting
	return composed.PatchStatus(backup).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeue).
		Run(ctx, state)
}
