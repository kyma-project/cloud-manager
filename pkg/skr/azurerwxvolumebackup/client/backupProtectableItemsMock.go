package client

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
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

	if ctx.Value("ListBackupProtectableItems") == "fail" {
		return result, errors.New("failed ListBackupProtectableItems")
	}

	if ctx.Value("ListBackupProtectableItems match") == 2 {
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

	if ctx.Value("ListBackupProtectableItems match") == 1 {

		if ctx.Value("NilName") == true {
			result = append(result, &armrecoveryservicesbackup.WorkloadProtectableItemResource{
				Name:       nil,
				Properties: &armrecoveryservicesbackup.AzureFileShareProtectableItem{FriendlyName: &friendlyName},
			},
			)

			return result, nil

		}

		// Continues on to CreateOrUpdateProtectedItem and TriggerBackup
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
