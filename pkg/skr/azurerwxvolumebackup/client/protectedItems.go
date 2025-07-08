package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

type ProtectedItemsClient interface {
	CreateOrUpdateProtectedItem(ctx context.Context, subscriptionId, location, vaultName, resourceGroupName, containerName, protectedItemName, backupPolicyName, storageAccountName string) error
	RemoveProtection(ctx context.Context, vaultName, resourceGroupName, containerName, protectedItemName string) error

	GetProtectedItem(ctx context.Context, protectedId string) (*armrecoveryservicesbackup.ProtectedItemResource, error)
	UpdateProtectedItem(ctx context.Context, protectedItem *armrecoveryservicesbackup.ProtectedItemResource) error
}

type protectedItemsClient struct {
	azureClient *armrecoveryservicesbackup.ProtectedItemsClient
}

func NewProtectedItemsClient(subscriptionId string, cred *azidentity.ClientSecretCredential) (ProtectedItemsClient, error) {

	pic, err := armrecoveryservicesbackup.NewProtectedItemsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	return protectedItemsClient{pic}, nil

}

// Binds a BackupPolicy to a Fileshare
func (c protectedItemsClient) CreateOrUpdateProtectedItem(ctx context.Context, subscriptionId, location, vaultName, resourceGroupName, containerName, protectedItemName, backupPolicyName, storageAccountName string) error {
	fabricName := "Azure"

	policyId := GetBackupPolicyPath(subscriptionId, resourceGroupName, vaultName, backupPolicyName)
	sourceResourceId := GetStorageAccountPath(subscriptionId, resourceGroupName, storageAccountName)

	parameters := armrecoveryservicesbackup.ProtectedItemResource{
		ETag:     nil,
		Location: to.Ptr(location),
		Properties: to.Ptr(armrecoveryservicesbackup.AzureFileshareProtectedItem{
			ProtectedItemType: to.Ptr("AzureFileShareProtectedItem"),
			PolicyID:          to.Ptr(policyId),
			SourceResourceID:  to.Ptr(sourceResourceId), // StorageAccountID
		}),
	}

	options := armrecoveryservicesbackup.ProtectedItemsClientCreateOrUpdateOptions{
		XMSAuthorizationAuxiliary: nil,
	}

	_, err := c.azureClient.CreateOrUpdate(ctx, vaultName, resourceGroupName, fabricName, containerName, protectedItemName, parameters, to.Ptr(options))
	if err != nil {
		return err
	}

	return nil

}

func (c protectedItemsClient) RemoveProtection(ctx context.Context, vaultName, resourceGroupName, containerName, protectedItemName string) error {
	fabricName := AzureFabricName
	_, err := c.azureClient.Delete(ctx, vaultName, resourceGroupName, fabricName, containerName, protectedItemName, nil)
	return err
}

func (c protectedItemsClient) GetProtectedItem(ctx context.Context, protectedId string) (*armrecoveryservicesbackup.ProtectedItemResource, error) {
	fabricName := AzureFabricName
	_, rgName, vaultName, containerName, name, err := ParseProtectedItemId(protectedId)

	if err != nil {
		return nil, err
	}

	protectedName := GetFileShareName(name)
	protected, err := c.azureClient.Get(ctx, vaultName, rgName, fabricName, containerName, protectedName, nil)
	if err != nil {
		return nil, err
	}

	return &protected.ProtectedItemResource, nil
}

func (c protectedItemsClient) UpdateProtectedItem(ctx context.Context, protectedItem *armrecoveryservicesbackup.ProtectedItemResource) error {

	if protectedItem == nil {
		return nil
	}

	fabricName := AzureFabricName
	_, rgName, vaultName, containerName, name, err := ParseProtectedItemId(*protectedItem.ID)
	if err != nil {
		return err
	}
	protectedName := GetFileShareName(name)

	options := armrecoveryservicesbackup.ProtectedItemsClientCreateOrUpdateOptions{
		XMSAuthorizationAuxiliary: nil,
	}

	_, err = c.azureClient.CreateOrUpdate(ctx, vaultName, rgName, fabricName, containerName, protectedName, *protectedItem, to.Ptr(options))
	if err != nil {
		return err
	}

	return err
}
