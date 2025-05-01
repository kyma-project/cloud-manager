package azurerwxvolumebackup

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func createBackup(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAzureRwxVolumeBackup()
	logger.WithValues("RwxBackup", backup.Name).Info("Creating Azure Rwx Volume Backup")

	// Setting the uuid as id to prevent duplicate backups
	if backup.Status.Id == "" {
		backup.Status.Id = uuid.NewString()

		return composed.UpdateStatus(backup).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)

	}

	vaultName := state.vaultName
	resourceGroupName := state.resourceGroupName

	storageAccountName := state.storageAccountName
	protectedItemName := state.protectedResourceName

	// Check if Fileshare is already protected. If it is, trigger backup and return

	// Bind BackupPolicy to Fileshare
	containerName := azurerwxvolumebackupclient.GetContainerName(resourceGroupName, storageAccountName)

	// Invoke backup
	err := state.client.TriggerBackup(ctx, vaultName, resourceGroupName, containerName, protectedItemName, backup.Spec.Location)
	if err != nil {

		logger.Error(err, "failed to trigger backup")
		logger.Info("retrying trigger backup")
		backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(
				metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonError,
					Message: fmt.Sprintf("failed to trigger backup: %s", err),
				},
			).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)

	}

	backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupDone // TODO: redo with Creating
	return composed.UpdateStatus(backup).SuccessError(composed.StopWithRequeueDelay(util.Timing.T60000ms())).Run(ctx, state)
}
