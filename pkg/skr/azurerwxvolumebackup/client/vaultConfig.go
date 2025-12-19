package client

import (
	"context"
	"errors"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

type VaultConfigClient interface {
	GetVaultConfig(ctx context.Context, resourceGroupName, vaultName string) (*armrecoveryservicesbackup.BackupResourceVaultConfigResource, error)
	PutVaultConfig(ctx context.Context, resourceGroupName, vaultName string, config *armrecoveryservicesbackup.BackupResourceVaultConfigResource) error
	GetStorageContainers(ctx context.Context, resourceGroupName, vaultName string) ([]*armrecoveryservicesbackup.ProtectionContainerResource, error)
	UnregisterContainer(ctx context.Context, resourceGroupName, vaultName, containerName string) error
}

type vaultConfigClient struct {
	configClient                     *armrecoveryservicesbackup.BackupResourceVaultConfigsClient
	backupProtectionContainersClient *armrecoveryservicesbackup.BackupProtectionContainersClient
	protectionContainersClient       *armrecoveryservicesbackup.ProtectionContainersClient
}

func NewVaultConfigClient(
	vc *armrecoveryservicesbackup.BackupResourceVaultConfigsClient,
	bpcc *armrecoveryservicesbackup.BackupProtectionContainersClient,
	pcc *armrecoveryservicesbackup.ProtectionContainersClient) VaultConfigClient {

	return vaultConfigClient{
		configClient:                     vc,
		backupProtectionContainersClient: bpcc,
		protectionContainersClient:       pcc,
	}
}

func (c vaultConfigClient) GetVaultConfig(ctx context.Context, resourceGroupName, vaultName string) (*armrecoveryservicesbackup.BackupResourceVaultConfigResource, error) {

	result, err := c.configClient.Get(
		ctx,
		vaultName,
		resourceGroupName,
		nil)

	if err != nil {
		return nil, err
	}

	return &result.BackupResourceVaultConfigResource, nil

}

func (c vaultConfigClient) PutVaultConfig(ctx context.Context, resourceGroupName, vaultName string, config *armrecoveryservicesbackup.BackupResourceVaultConfigResource) error {

	if config == nil {
		return errors.New("vault config cannot be nil")
	}

	_, err := c.configClient.Put(
		ctx,
		vaultName,
		resourceGroupName,
		*config,
		nil)

	return err
}

func (c vaultConfigClient) GetStorageContainers(ctx context.Context, resourceGroupName, vaultName string) ([]*armrecoveryservicesbackup.ProtectionContainerResource, error) {
	pager := c.backupProtectionContainersClient.NewListPager(vaultName,
		resourceGroupName,
		&armrecoveryservicesbackup.BackupProtectionContainersClientListOptions{
			Filter: to.Ptr("backupManagementType eq 'AzureStorage'"),
		},
	)

	var containers []*armrecoveryservicesbackup.ProtectionContainerResource
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return containers, err
		}

		containers = append(containers, page.Value...)

	}
	return containers, nil
}

func (c vaultConfigClient) UnregisterContainer(ctx context.Context, resourceGroupName, vaultName, containerName string) error {
	_, err := c.protectionContainersClient.Unregister(
		ctx,
		vaultName,
		resourceGroupName,
		AzureFabricName,
		containerName,
		nil,
	)

	return err
}
