package client

import (
	"context"
	"errors"
)

func newProtectedItemsMockClient() *protectedItemsClient {
	return &protectedItemsClient{}
}

type protectedItemsMockClient struct {
	protectedItemsClient
}

func (m *protectedItemsMockClient) CreateOrUpdateProtectedItem(ctx context.Context,
	subscriptionId,
	location,
	vaultName,
	resourceGroupName,
	containerName,
	protectedItemName,
	backupPolicyName,
	storageAccountName string) error {

	if vaultName == "vaultName - one fail CreateOrUpdateProtectedItem" {
		return errors.New("failed CreateOrUpdateProtectedItem")
	}

	return nil
}
