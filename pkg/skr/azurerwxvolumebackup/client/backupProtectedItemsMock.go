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

	if vaultName == "fail ListProtectedItems" {
		return result, errors.New("failed ListProtectedItems")
	}

	if vaultName == "more than 1" {
		friendlyName := "matchingFileShareName"
		name := "ProtectedItemName"
		protectedItems := []*armrecoveryservicesbackup.ProtectedItemResource{
			{Name: &name, Properties: &armrecoveryservicesbackup.AzureFileshareProtectedItem{FriendlyName: &friendlyName}},
			{Name: &name, Properties: &armrecoveryservicesbackup.AzureFileshareProtectedItem{FriendlyName: &friendlyName}},
		}

		return protectedItems, nil

	}

	if vaultName == "exactly 1 - fail" || vaultName == "exactly 1 - succeed" {
		friendlyName := "matchingFileShareName"
		name := "ProtectedItemName"
		protectedItems := []*armrecoveryservicesbackup.ProtectedItemResource{
			{Name: &name, Properties: &armrecoveryservicesbackup.AzureFileshareProtectedItem{FriendlyName: &friendlyName}},
		}

		return protectedItems, nil

	}

	// happy path
	return result, nil

}
