package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"k8s.io/utils/ptr"

	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
)

const (
	AzureRecoveryVault       = "AzureRecoveryVault"
	AzureStorageContainer    = "AzureStorageContainer"
	AzureFileShareProtection = "AzureFileShareProtection"
)

type NukeRwxBackupClient interface {
	ListRwxVolumeBackupVaults(ctx context.Context) ([]*armrecoveryservices.Vault, error)
	ListFileShareProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (map[string]*armrecoveryservicesbackup.AzureFileshareProtectedItem, error)
	ListStorageContainers(ctx context.Context, vault *armrecoveryservices.Vault) (map[string]*armrecoveryservicesbackup.AzureStorageContainer, error)
	HasProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error)
	HasProtectionContainers(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error)
	DisableSoftDelete(ctx context.Context, vault *armrecoveryservices.Vault) error
	RemoveProtection(ctx context.Context, protectedId string) error
	UnregisterContainer(ctx context.Context, containerId string) error
	DeleteVault(ctx context.Context, vault *armrecoveryservices.Vault) error
}

type nukeRwxBackupClient struct {
	azurerwxvolumebackupclient.Client
}

func NewClientProvider() azureclient.ClientProvider[NukeRwxBackupClient] {
	return NukeProvider(azurerwxvolumebackupclient.NewClientProvider())
}

func NukeProvider(backupProvider azureclient.ClientProvider[azurerwxvolumebackupclient.Client]) azureclient.ClientProvider[NukeRwxBackupClient] {

	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (NukeRwxBackupClient, error) {
		client, err := backupProvider(ctx, clientId, clientSecret, subscriptionId, tenantId, auxiliaryTenants...)

		if err != nil {
			return nil, err
		}

		return &nukeRwxBackupClient{Client: client}, nil
	}
}

func (c *nukeRwxBackupClient) ListRwxVolumeBackupVaults(ctx context.Context) ([]*armrecoveryservices.Vault, error) {

	var result []*armrecoveryservices.Vault

	//Get all the vaults associated with this subscription
	vaults, err := c.ListVaults(ctx)
	if err != nil {
		return result, err
	}

	for _, vault := range vaults {
		if value, exists := vault.Tags[azurerwxvolumebackupclient.TagNameCloudManager]; exists && ptr.Deref(value, "") == azurerwxvolumebackupclient.TagValueRwxVolumeBackup {
			result = append(result, vault)
		}
	}

	return result, nil
}

func (c *nukeRwxBackupClient) ListFileShareProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (map[string]*armrecoveryservicesbackup.AzureFileshareProtectedItem, error) {
	result := make(map[string]*armrecoveryservicesbackup.AzureFileshareProtectedItem)

	if vault == nil {
		return result, nil
	}

	_, rgName, vaultName, err := azurerwxvolumebackupclient.ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return result, nil
	}

	protectedItems, err := c.ListProtectedItems(ctx, vaultName, rgName)
	if err != nil {
		return result, nil
	}

	for _, item := range protectedItems {
		switch protected := item.Properties.(type) {
		case *armrecoveryservicesbackup.AzureFileshareProtectedItem:
			result[*item.ID] = protected
		default:
			continue
		}
	}

	return result, nil

}

func (c *nukeRwxBackupClient) ListStorageContainers(ctx context.Context, vault *armrecoveryservices.Vault) (map[string]*armrecoveryservicesbackup.AzureStorageContainer, error) {
	result := make(map[string]*armrecoveryservicesbackup.AzureStorageContainer)

	if vault == nil {
		return result, nil
	}

	_, rgName, vaultName, err := azurerwxvolumebackupclient.ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return result, nil
	}

	protectedItems, err := c.GetStorageContainers(ctx, rgName, vaultName)
	if err != nil {
		return result, nil
	}

	for _, item := range protectedItems {
		switch protected := item.Properties.(type) {
		case *armrecoveryservicesbackup.AzureStorageContainer:
			result[*item.ID] = protected
		default:
			continue
		}
	}

	return result, nil
}

func (c *nukeRwxBackupClient) HasProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error) {
	if vault == nil {
		return false, nil
	}

	_, rgName, vaultName, err := azurerwxvolumebackupclient.ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return false, err
	}

	protectedItems, err := c.ListProtectedItems(ctx, vaultName, rgName)
	if err != nil {
		return false, nil
	}

	return len(protectedItems) > 0, nil
}

func (c *nukeRwxBackupClient) HasProtectionContainers(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error) {
	if vault == nil {
		return false, nil
	}

	_, rgName, vaultName, err := azurerwxvolumebackupclient.ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return false, err
	}

	protectedItems, err := c.GetStorageContainers(ctx, rgName, vaultName)
	if err != nil {
		return false, nil
	}

	return len(protectedItems) > 0, nil
}

func (c *nukeRwxBackupClient) DisableSoftDelete(ctx context.Context, vault *armrecoveryservices.Vault) error {
	if vault == nil {
		return nil
	}

	_, rgName, vaultName, err := azurerwxvolumebackupclient.ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return err
	}

	config, err := c.GetVaultConfig(ctx, rgName, vaultName)
	if err != nil {
		return err
	}

	config.Properties.SoftDeleteFeatureState = ptr.To(armrecoveryservicesbackup.SoftDeleteFeatureStateDisabled)

	return c.PutVaultConfig(ctx, rgName, vaultName, config)
}

func (c *nukeRwxBackupClient) RemoveProtection(ctx context.Context, protectedId string) error {

	_, rgName, vaultName, containerName, protectedName, err := azurerwxvolumebackupclient.ParseProtectedItemId(protectedId)
	if err != nil {
		return err
	}

	//Remove Protection
	fileShareName := azurerwxvolumebackupclient.GetFileShareName(protectedName)
	err = c.Client.RemoveProtection(ctx, vaultName, rgName, containerName, fileShareName)
	if err != nil {
		return err
	}

	return nil
}

func (c *nukeRwxBackupClient) UnregisterContainer(ctx context.Context, containerId string) error {

	_, rgName, vaultName, containerName, err := azurerwxvolumebackupclient.ParseContainerId(containerId)
	if err != nil {
		return err
	}

	//Unregister the containers
	err = c.Client.UnregisterContainer(ctx, rgName, vaultName, containerName)
	if err != nil {
		return err
	}

	return nil
}

func (c *nukeRwxBackupClient) DeleteVault(ctx context.Context, vault *armrecoveryservices.Vault) error {
	if vault == nil {
		return nil
	}

	_, rgName, vaultName, err := azurerwxvolumebackupclient.ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return err
	}

	err = c.Client.DeleteVault(ctx, rgName, vaultName)
	return err
}
