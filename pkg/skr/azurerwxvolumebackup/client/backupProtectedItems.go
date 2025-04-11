package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

type BackupProtectedItemsClient interface {
	ListProtectedItems(ctx context.Context, vaultName, resourceGroupName string) ([]*armrecoveryservicesbackup.ProtectedItemResource, error)
}

type backupProtectedItemsClient struct {
	azureClient *armrecoveryservicesbackup.BackupProtectedItemsClient
}

func NewBackupProtectedItemsClient(subscriptionId string, cred *azidentity.ClientSecretCredential) (BackupProtectedItemsClient, error) {

	c, err := armrecoveryservicesbackup.NewBackupProtectedItemsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	return backupProtectedItemsClient{c}, nil

}

func (c backupProtectedItemsClient) ListProtectedItems(ctx context.Context, vaultName, resourceGroupName string) ([]*armrecoveryservicesbackup.ProtectedItemResource, error) {

	var result []*armrecoveryservicesbackup.ProtectedItemResource

	pager := c.azureClient.NewListPager(vaultName, resourceGroupName, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {

			return result, err

		}

		result = append(result, page.Value...)
	}

	return result, nil

}
