package client

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"strings"
)

func newBackupProtectableItemsMockClient() *backupProtectableItemsClient {
	return &backupProtectableItemsClient{}
}

type backupProtectableItemsMockClient struct {
	backupProtectableItemsClient
}

func (m *backupProtectableItemsMockClient) ListBackupProtectableItems(ctx context.Context, vaultName string, resourceGroupName string) ([]*armrecoveryservicesbackup.WorkloadProtectableItemResource, error) {

	var result []*armrecoveryservicesbackup.WorkloadProtectableItemResource
	var friendlyName = "matchingFileShareName"

	if vaultName == "vaultName - fail ListBackupProtectableItems" {
		return result, errors.New("failed ListBackupProtectableItems")
	}

	if vaultName == "vaultName - more than 1" {
		name := "vaultName - more than 1"
		result = append(result, &armrecoveryservicesbackup.WorkloadProtectableItemResource{
			Name:       &name,
			Properties: &armrecoveryservicesbackup.AzureFileShareProtectableItem{FriendlyName: &friendlyName},
		}, &armrecoveryservicesbackup.WorkloadProtectableItemResource{
			Name:       &name,
			Properties: &armrecoveryservicesbackup.AzureFileShareProtectableItem{FriendlyName: &friendlyName},
		})

		return result, nil

	}

	if vaultName == "vaultName - one nil" {
		result = append(result, &armrecoveryservicesbackup.WorkloadProtectableItemResource{
			Name:       nil,
			Properties: &armrecoveryservicesbackup.AzureFileShareProtectableItem{FriendlyName: &friendlyName},
		},
		)

		return result, nil

	}

	// Continues on to CreateOrUpdateProtectedItem and TriggerBackup
	if strings.HasPrefix(vaultName, "vaultName - one") {

		//if vaultName == "vaultName - one" || vaultName == "vaultName - one fail CreateOrUpdateProtectedItem" {
		name := "vaultName - one"
		result = append(result, &armrecoveryservicesbackup.WorkloadProtectableItemResource{
			Name:       &name,
			Properties: &armrecoveryservicesbackup.AzureFileShareProtectableItem{FriendlyName: &friendlyName},
		},
		)

		return result, nil

	}

	return result, nil

}
