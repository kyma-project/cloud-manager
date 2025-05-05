package azurerwxvolumebackup

import (
	"context"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
)

func unbindBackupPolicy(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAzureRwxVolumeBackup()
	logger.WithValues("RwxBackup", backup.Name).Info("Deleting Azure Rwx Volume Backup")

	if !composed.MarkedForDeletionPredicate(ctx, st) {
		return nil, ctx
	}

	// Fetch variables from state
	vaultName := state.vaultName
	resourceGroupName := state.resourceGroupName
	storageAccount := state.storageAccountName
	containerName := client.GetContainerName(resourceGroupName, storageAccount)

	// TODO: implement fetching via action
	protectedResourceName := state.protectedResourceName

	err := state.client.RemoveProtection(ctx, vaultName, resourceGroupName, containerName, protectedResourceName)
	if err != nil {
		backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupError
		return composed.PatchStatus(backup).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	return nil, ctx

}
