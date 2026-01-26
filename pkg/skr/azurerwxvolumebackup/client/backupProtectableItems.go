package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

type BackupProtectableItemsClient interface {
	ListBackupProtectableItems(ctx context.Context, vaultName string, resourceGroupName string) ([]*armrecoveryservicesbackup.WorkloadProtectableItemResource, error)
}

type backupProtectableItemsClient struct {
	azureClient *armrecoveryservicesbackup.BackupProtectableItemsClient
}

func NewBackupProtectableItemsClient(bpic *armrecoveryservicesbackup.BackupProtectableItemsClient) BackupProtectableItemsClient {
	return backupProtectableItemsClient{bpic}
}

// Lists out Fileshares that can be bound by a BackupPolicy
func (c backupProtectableItemsClient) ListBackupProtectableItems(ctx context.Context, vaultName string, resourceGroupName string) ([]*armrecoveryservicesbackup.WorkloadProtectableItemResource, error) {

	pager := c.azureClient.NewListPager(vaultName, resourceGroupName,
		to.Ptr(armrecoveryservicesbackup.BackupProtectableItemsClientListOptions{
			Filter:    to.Ptr("backupManagementType eq 'AzureStorage'"),
			SkipToken: nil,
		}),
	)

	var res []*armrecoveryservicesbackup.WorkloadProtectableItemResource
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return res, err
		}
		res = append(res, page.Value...)
	}
	return res, nil

}
