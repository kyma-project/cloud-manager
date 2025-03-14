package azurerwxvolumebackup

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	path "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	"github.com/kyma-project/cloud-manager/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type protectedFileShare struct {
	Id         *string
	Name       *string
	Properties *armrecoveryservicesbackup.AzureFileshareProtectedItem
}

func createBackup(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAzureRwxVolumeBackup()
	logger.WithValues("RwxBackup", backup.Name).Info("Creating Azure Rwx Volume Backup")

	// UUID; early return
	if backup.Status.Id == "" {
		// get location?
		backup.Status.Id = uuid.NewString()
		return composed.PatchStatus(backup).
			SetExclusiveConditions().
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	// TODO: figure out where we get the vaultName, policyName
	vaultName := "vaultName"
	resourceGroupName := state.resourceGroupName
	fileShareName := state.fileShareName
	policyName := "policyName"

	// TODO: figure out subscriptionId
	subscriptionId := "subscriptionId"
	storageAccountName := state.storageAccountName

	// Check if Fileshare is already protected. If it is, trigger backup and return
	protectedItems, err := state.client.ListProtectedItems(ctx, vaultName, resourceGroupName)
	var protectedItemsMatching []protectedFileShare

	for _, item := range protectedItems {
		iProperties, ok := item.Properties.(*armrecoveryservicesbackup.AzureFileshareProtectedItem)
		if !ok {
			continue
		}

		if iProperties.FriendlyName != nil && *iProperties.FriendlyName == fileShareName {
			protectedItemsMatching = append(protectedItemsMatching, protectedFileShare{
				Id:         item.ID,
				Name:       item.Name,
				Properties: iProperties,
			})
		}

	}

	if len(protectedItemsMatching) > 1 {
		// error, there's more than 1 protectedItemsMatching
		return composed.PatchStatus(backup).
			SetExclusiveConditions().
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	if len(protectedItemsMatching) == 1 {
		// Invoke backup and return
		protectedItemName := *protectedItemsMatching[0].Name
		containerName := path.GetContainerName(resourceGroupName, storageAccountName)
		err = state.client.TriggerBackup(ctx, vaultName, resourceGroupName, containerName, protectedItemName, backup.Spec.Location)
		if err != nil {
			return composed.PatchStatus(backup).
				SetExclusiveConditions(). // TODO
				SuccessError(composed.StopWithRequeue).
				Run(ctx, state)
		}
		return composed.PatchStatus(backup).
			SetExclusiveConditions(). // TODO
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	// Case: there's no protected items matching
	// Create BackupPolicy on Fileshare
	err = state.client.CreateBackupPolicy(ctx, vaultName, resourceGroupName, policyName)
	if err != nil {
		logger.Error(err, "failed to create backup policy")
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: fmt.Sprintf("Cloud not create BackupPolicy for backup: %s", err),
			}).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	// Fetch unfriendly name of unprotected fileshare
	protectableItems, err := state.client.ListBackupProtectableItems(ctx, vaultName, resourceGroupName)
	if err != nil {
		// handle error & return
		logger.Error(err, "failed to fetch protectable items")
		return composed.PatchStatus(backup).
			SetExclusiveConditions(). // TODO
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	var matchingItems []*armrecoveryservicesbackup.WorkloadProtectableItemResource
	for _, item := range protectableItems {

		props, ok := item.Properties.(*armrecoveryservicesbackup.AzureFileShareProtectableItem)
		if !ok || props == nil {
			logger.Error(err, "failed to type cast to *armrecoveryservicesbackup.AzureFileShareProtectableItem")

			return composed.PatchStatus(backup).
				SetExclusiveConditions(). // TODO
				SuccessError(composed.StopAndForget).
				Run(ctx, state)

		}

		// Skip over entries that don't have friendly name
		// Match FriendlyName
		if props.FriendlyName != nil && *props.FriendlyName == fileShareName {
			matchingItems = append(matchingItems, item)
		}

	}

	if len(matchingItems) == 0 {
		logger.Error(err, "fileshare not found in ListBackupProtectableItems")
		logger.Info("requeue create backup")
		return composed.PatchStatus(backup).
			SetExclusiveConditions(). // TODO
			// Give some time for Fileshare to show up as Protectable
			// Try again in 5 minutes
			SuccessError(composed.StopWithRequeueDelay(3e11)).
			Run(ctx, state)

	}

	if len(matchingItems) > 1 {
		logger.Error(err, "more than 1 friendlyName found")
		return composed.PatchStatus(backup).
			SetExclusiveConditions(). // TODO
			SuccessError(composed.StopAndForget).
			Run(ctx, state)

	}

	if matchingItems[0].Name == nil {
		logger.Error(err, "matching item's Name is nil")
		return composed.PatchStatus(backup).
			SetExclusiveConditions(). // TODO
			SuccessError(composed.StopAndForget).
			Run(ctx, state)
	}

	// Bind BackupPolicy to Fileshare
	protectedItemName := *matchingItems[0].Name
	containerName := path.GetContainerName(resourceGroupName, storageAccountName)
	err = state.client.CreateOrUpdateProtectedItem(ctx, subscriptionId, "location", vaultName, resourceGroupName, containerName, protectedItemName, policyName, storageAccountName)
	if err != nil {

		logger.Error(err, "failed to bind backup policy to fileshare")
		logger.Info("retrying binding backup policy to fileshare")
		return composed.PatchStatus(backup).
			SetExclusiveConditions(). // TODO
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)

	}

	// Invoke backup
	err = state.client.TriggerBackup(ctx, vaultName, resourceGroupName, containerName, protectedItemName, backup.Spec.Location)
	if err != nil {

		logger.Error(err, "failed to trigger backup")
		logger.Info("retrying trigger backup")
		return composed.PatchStatus(backup).
			SetExclusiveConditions(). // TODO
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)

	}

	return composed.StopWithRequeueDelay(util.Timing.T60000ms()), nil
}
