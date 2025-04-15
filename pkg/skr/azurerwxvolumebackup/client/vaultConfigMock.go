package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

func newVaultConfigMockClient() VaultConfigClient {
	return &vaultConfigMockClient{}
}

type vaultConfigMockClient struct {
}

func (m *vaultConfigMockClient) GetVaultConfig(ctx context.Context, resourceGroupName, vaultName string) (*armrecoveryservicesbackup.BackupResourceVaultConfigResource, error) {

	// happy path
	config := armrecoveryservicesbackup.BackupResourceVaultConfigResource{}
	return &config, nil

}

func (m *vaultConfigMockClient) PutVaultConfig(ctx context.Context, resourceGroupName, vaultName string, config *armrecoveryservicesbackup.BackupResourceVaultConfigResource) error {

	// happy path
	return nil

}
func (m *vaultConfigMockClient) GetStorageContainers(ctx context.Context, resourceGroupName, vaultName string) ([]*armrecoveryservicesbackup.ProtectionContainerResource, error) {
	return nil, nil
}

func (m *vaultConfigMockClient) UnregisterContainer(ctx context.Context, resourceGroupName, vaultName, containerName string) error {

	// happy path
	return nil
}
