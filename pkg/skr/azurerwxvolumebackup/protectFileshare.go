package azurerwxvolumebackup

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type protectableFileShare struct {
	Id         *string
	Name       *string
	Properties *armrecoveryservicesbackup.AzureFileShareProtectableItem
}

func matchProtectableItems(protectableItems []*armrecoveryservicesbackup.WorkloadProtectableItemResource, fileShareName string) []protectableFileShare {

	//var matchingItems []*armrecoveryservicesbackup.WorkloadProtectableItemResource
	var matchingItems []protectableFileShare
	for _, item := range protectableItems {

		props, ok := item.Properties.(*armrecoveryservicesbackup.AzureFileShareProtectableItem)
		if !ok || props == nil {
			continue
		}

		// Skip over entries that don't have friendly name
		// Match FriendlyName
		if props.FriendlyName != nil && *props.FriendlyName == fileShareName {
			matchingItems = append(matchingItems, protectableFileShare{
				Id:         item.ID,
				Name:       item.Name,
				Properties: props,
			})
		}

	}
	return matchingItems

}

func protectFileshare(ctx context.Context, st composed.State) (error, context.Context) {

	state := st.(*State)
	logger := composed.LoggerFromCtx(ctx)
	backup := state.ObjAsAzureRwxVolumeBackup()

	// Short circuit if state.protectedResourceName  not nil
	if state.protectedResourceName != "" {
		return nil, ctx
	}

	fileShareName := state.fileShareName
	vaultName := state.vaultName
	resourceGroupName := state.resourceGroupName
	storageAccountName := state.storageAccountName
	subscriptionId := state.subscriptionId

	// Fetch unfriendly name of unprotected fileshare
	protectableItems, err := state.client.ListBackupProtectableItems(ctx, vaultName, resourceGroupName)
	if err != nil {
		// handle error & return
		logger.Error(err, "failed to fetch protectable items")
		backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(
				metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonError,
					Message: fmt.Sprintf("Could not fetch Backup Protectable Items: %s", err),
				},
			).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	matchingItems := matchProtectableItems(protectableItems, fileShareName)

	if len(matchingItems) == 0 {
		// stop and requeue; give some time
		logger.Error(errors.New("fileshare not found in ListBackupProtectableItems"), "fileshare not found in ListBackupProtectableItems")
		logger.Info("requeue create backup")
		backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(metav1.Condition{
				Type:    cloudresourcesv1beta1.ConditionTypeError,
				Status:  metav1.ConditionTrue,
				Reason:  cloudresourcesv1beta1.ConditionReasonError,
				Message: "fileshare not found in ListBackupProtectableItems",
			}).
			// Give some time for Fileshare to show up as Protectable
			// Try again in 5 minutes
			SuccessError(composed.StopWithRequeueDelay(3e11)).
			Run(ctx, state)

	}

	if len(matchingItems) > 1 {
		logger.Error(err, "more than 1 friendlyName found")
		backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupFailed
		return composed.PatchStatus(backup).
			SetExclusiveConditions(
				metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonError,
					Message: "more than one friendlyName found",
				},
			).
			FailedError(composed.StopAndForget).
			Run(ctx, state)
	}

	if matchingItems[0].Name == nil {
		logger.Error(errors.New("matching item's Name is nil"), "matching item's Name is nil")
		backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupFailed
		return composed.PatchStatus(backup).
			SetExclusiveConditions(
				metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonError,
					Message: "matching item's Name is nil",
				},
			).
			FailedError(composed.StopAndForget).
			Run(ctx, state)
	}

	protectableItemName := *matchingItems[0].Name
	containerName := azurerwxvolumebackupclient.GetContainerName(resourceGroupName, storageAccountName)
	location := backup.Spec.Location

	err = state.client.CreateOrUpdateProtectedItem(ctx, subscriptionId, location, vaultName, resourceGroupName, containerName, protectableItemName, azurerwxvolumebackupclient.DefaultBackupPolicyName, storageAccountName)
	if err != nil {
		logger.Error(err, "failed to bind backup policy to fileshare")
		logger.Info("retrying binding backup policy to fileshare")
		backup.Status.State = cloudresourcesv1beta1.AzureRwxBackupError
		return composed.PatchStatus(backup).
			SetExclusiveConditions(
				metav1.Condition{
					Type:    cloudresourcesv1beta1.ConditionTypeError,
					Status:  metav1.ConditionTrue,
					Reason:  cloudresourcesv1beta1.ConditionReasonError,
					Message: "failed to bind backup policy to fileshare",
				},
			).
			SuccessError(composed.StopWithRequeue).
			Run(ctx, state)
	}

	state.protectedResourceName = protectableItemName
	return nil, ctx

}
