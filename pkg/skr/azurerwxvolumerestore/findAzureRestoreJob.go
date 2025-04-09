package azurerwxvolumerestore

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"time"
)

func findAzureRestoreJob(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsAzureRwxVolumeRestore()
	if restore.Status.OpIdentifier != "" {
		return nil, ctx
	}
	if restore.Status.RestoredDir == "" {
		// Restore was not tried yet
		return nil, ctx
	}
	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Finding restore job ID in case restore already started but opIdentifier failed to be set")
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

	jobId, retry, err := state.storageClient.FindRestoreJobId(ctx, vault, resourceGroup, state.fileShareName, restore.Status.StartTime.Format(time.RFC3339), restore.Status.RestoredDir)
	if err != nil {
		return composed.LogErrorAndReturn(err, "Error when trying to find restore job ID", composed.StopAndForget, ctx)
	}
	if retry {
		logger.Info("Restore job not found, retrying later")
		return composed.StopWithRequeueDelay(util.Timing.T10000ms()), ctx
	}
	if jobId == nil {
		return nil, ctx
	}
	restore.Status.OpIdentifier = *jobId
	return composed.PatchStatus(restore).SuccessErrorNil().Run(ctx, state)
}
