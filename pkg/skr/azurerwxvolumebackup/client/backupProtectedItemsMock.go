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

	if ctx.Value("ListProtectedItems") == "fail" {
		return result, errors.New("failed ListProtectedItems")
	}

	if ctx.Value("ListProtectedItems match") == 2 {

		protectedItems := []*armrecoveryservicesbackup.ProtectedItemResource{
			{Name: &protectedItemName, Properties: &armrecoveryservicesbackup.AzureFileshareProtectedItem{FriendlyName: &friendlyName}},
			{Name: &protectedItemName, Properties: &armrecoveryservicesbackup.AzureFileshareProtectedItem{FriendlyName: &friendlyName}},
		}

		return protectedItems, nil

	}

	if ctx.Value("ListProtectedItems match") == 1 {
		protectedItems := []*armrecoveryservicesbackup.ProtectedItemResource{
			{Name: &protectedItemName, Properties: &armrecoveryservicesbackup.AzureFileshareProtectedItem{FriendlyName: &friendlyName}},
		}

		return protectedItems, nil

	}

	// happy path
	return result, nil

}
