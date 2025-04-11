package azurerwxvolumerestore

import (
	"context"
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func startAzureRestore(ctx context.Context, st composed.State) (error, context.Context) {
	state := st.(*State)
	restore := state.ObjAsAzureRwxVolumeRestore()
	if restore.Status.OpIdentifier != "" {
		return nil, ctx
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.Info("Starting Azure Restore")
	_, resourceGroup, vault, container, protectedItem, recoveryPointId, err := client.ParseRecoveryPointId(state.azureRwxVolumeBackup.Status.RecoveryPointId)
	if err != nil {
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		errorMessage := fmt.Sprintf("Source AzureRwxVolumeBackup has an invalid recoveryPointId: '%v'", state.azureRwxVolumeBackup.Status.RecoveryPointId)
		logger.Error(err, errorMessage)
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonInvalidRecoveryPointId,
				Message: errorMessage,
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}
	sourceSAPath := state.azureRwxVolumeBackup.Status.StorageAccountPath
	if sourceSAPath == "" {
		restore.Status.State = cloudresourcesv1beta1.JobStateFailed
		errorMessage := "Source AzureRwxVolumeBackup has an empty storageAccountPath"
		logger.Error(nil, errorMessage)
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonInvalidStorageAccountPath,
				Message: errorMessage,
			}).
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}
	targetSAPath := client.GetStorageAccountPath(state.Scope().Spec.Scope.Azure.SubscriptionId, state.resourceGroupName, state.storageAccountName)
	restoreRequest := client.RestoreRequest{
		VaultName:                vault,
		ResourceGroupName:        resourceGroup,
		FabricName:               client.AzureFabricName,
		ContainerName:            container,
		ProtectedItemName:        protectedItem,
		RecoveryPointId:          recoveryPointId,
		SourceStorageAccountPath: sourceSAPath,
		TargetStorageAccountPath: targetSAPath,
		TargetFileShareName:      state.fileShareName,
		TargetFolderName:         restore.Status.RestoredDir,
	}
	jobId, err := state.storageClient.TriggerRestore(ctx, restoreRequest)

	if err != nil {
		restore.Status.State = cloudresourcesv1beta1.JobStateError
		logger.Error(err, "Error starting Azure Restore")
		restore.Status.StartTime = nil
		restore.Status.RestoredDir = ""
		return composed.PatchStatus(restore).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonErrorStartingRestore,
				Message: "Error starting Azure Restore",
			}).
			SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
			Run(ctx, state)
	}

	//Update the status with opIdentifier and InProgress state.
	restore.Status.OpIdentifier = ptr.Deref(jobId, "")
	restore.Status.State = cloudresourcesv1beta1.JobStateInProgress
	return composed.PatchStatus(restore).
		SetExclusiveConditions().
		SuccessError(composed.StopWithRequeueDelay(util.Timing.T1000ms())).
		Run(ctx, state)
}
