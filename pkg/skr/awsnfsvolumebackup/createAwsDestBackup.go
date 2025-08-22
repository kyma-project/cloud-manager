package awsnfsvolumebackup

import (
	"context"
	"fmt"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func createAwsDestBackup(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAwsNfsVolumeBackup()

	//If the object is being deleted, continue...
	if composed.IsMarkedForDeletion(backup) {
		return nil, ctx
	}

	//If the aws destination backup was already created, continue...
	if len(backup.Status.CopyJobId) > 0 || len(backup.Status.RemoteId) > 0 {
		return nil, ctx
	}

	// Backup does not exist
	logger.Info("Creating AWS Destination Backup")

	//Create a Backup Job
	res, err := state.awsClient.StartCopyJob(ctx, &client.StartCopyJobInput{
		SourceBackupVaultName:     state.GetVaultName(),
		DestinationBackupVaultArn: state.GetDestinationBackupVaultArn(),
		RecoveryPointArn:          state.GetRecoveryPointArn(),
		IamRoleArn:                state.GetBackupRoleArn(),
		IdempotencyToken:          ptr.To(backup.Status.IdempotencyToken),
	})
	if err != nil {
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionTypeError,
				Message: err.Error(),
			}).
			SuccessLogMsg(fmt.Sprintf("Error creating AWS destination backup : %s", err)).
			SuccessError(err).
			Run(ctx, state)
	}

	//Update the status with details.
	backup.Status.CopyJobId = ptr.Deref(res.CopyJobId, "")
	backup.Status.State = cloudresourcesv1beta1.StateCreatingRemote
	return composed.PatchStatus(backup).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
		Run(ctx, state)
}
