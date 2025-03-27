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

	if vaultName == "vaultName - fail ListBackupProtectableItems" {
		return result, errors.New("failed ListBackupProtectableItems")
	}

	return result, nil

}
