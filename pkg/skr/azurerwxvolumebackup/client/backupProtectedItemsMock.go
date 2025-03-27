package client

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

func newBackupProtectedItemsMockClient() *backupProtectedItemsClient {
	return &backupProtectedItemsClient{}
}

type backupProtectedItemsMockClient struct {
	backupProtectedItemsClient
}

func (m *backupProtectedItemsMockClient) ListProtectedItems(ctx context.Context, vaultName, resourceGroupName string) ([]*armrecoveryservicesbackup.ProtectedItemResource, error) {

	var result []*armrecoveryservicesbackup.ProtectedItemResource
	var friendlyName = "matchingFileShareName"
	var protectedItemName = "ProtectedItemName"

	if vaultName == "fail ListProtectedItems" {
		return result, errors.New("failed ListProtectedItems")
	}

	if vaultName == "more than 1" {

		protectedItems := []*armrecoveryservicesbackup.ProtectedItemResource{
			{Name: &protectedItemName, Properties: &armrecoveryservicesbackup.AzureFileshareProtectedItem{FriendlyName: &friendlyName}},
			{Name: &protectedItemName, Properties: &armrecoveryservicesbackup.AzureFileshareProtectedItem{FriendlyName: &friendlyName}},
		}

		return protectedItems, nil

	}

	if vaultName == "exactly 1 - fail" || vaultName == "exactly 1 - succeed" {
		protectedItems := []*armrecoveryservicesbackup.ProtectedItemResource{
			{Name: &protectedItemName, Properties: &armrecoveryservicesbackup.AzureFileshareProtectedItem{FriendlyName: &friendlyName}},
		}

		return protectedItems, nil

	}

	// happy path
	return result, nil

}
