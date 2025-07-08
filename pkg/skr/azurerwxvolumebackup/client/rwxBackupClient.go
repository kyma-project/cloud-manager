package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"k8s.io/utils/ptr"
)

type RwxBackupClient interface {
	ListRwxVolumeBackupVaults(ctx context.Context, shootName string) ([]*armrecoveryservices.Vault, error)
	ListFileShareProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (map[string]*armrecoveryservicesbackup.AzureFileshareProtectedItem, error)
	ListStorageContainers(ctx context.Context, vault *armrecoveryservices.Vault) (map[string]*armrecoveryservicesbackup.AzureStorageContainer, error)
	HasProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error)
	HasProtectionContainers(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error)
	DisableSoftDelete(ctx context.Context, vault *armrecoveryservices.Vault) error
	StopFileShareProtection(ctx context.Context, protectedId string, protected *armrecoveryservicesbackup.AzureFileshareProtectedItem) error
	RemoveProtection(ctx context.Context, protectedId string) error
	UnregisterContainer(ctx context.Context, containerId string) error
	DeleteVault(ctx context.Context, vault *armrecoveryservices.Vault) error
}

type rwxBackupClient struct {
	Client
}

func NewRwxBackupClientProvider() azureclient.ClientProvider[RwxBackupClient] {
	return RwxBackupClientProvider(NewClientProvider())
}

func RwxBackupClientProvider(backupClientProvider azureclient.ClientProvider[Client]) azureclient.ClientProvider[RwxBackupClient] {

	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (RwxBackupClient, error) {
		client, err := backupClientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId, auxiliaryTenants...)

		if err != nil {
			return nil, err
		}

		return &rwxBackupClient{Client: client}, nil
	}
}

func (c *rwxBackupClient) ListRwxVolumeBackupVaults(ctx context.Context, shootName string) ([]*armrecoveryservices.Vault, error) {

	var result []*armrecoveryservices.Vault

	//Get all the vaults associated with this subscription
	vaults, err := c.ListVaults(ctx)
	if err != nil {
		return result, err
	}

	for _, vault := range vaults {
		_, rgName, _, err := ParseVaultId(ptr.Deref(vault.ID, ""))
		if err != nil {
			composed.LoggerFromCtx(ctx).Error(err, fmt.Sprintf("Error parsing vault ID: %v", *vault.ID))
			continue
		}
		if !strings.HasSuffix(rgName, shootName) {
			continue
		}
		if value, exists := vault.Tags[TagNameCloudManager]; exists && ptr.Deref(value, "") == TagValueRwxVolumeBackup {
			result = append(result, vault)
		}
	}

	return result, nil
}

func (c *rwxBackupClient) ListFileShareProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (map[string]*armrecoveryservicesbackup.AzureFileshareProtectedItem, error) {
	result := make(map[string]*armrecoveryservicesbackup.AzureFileshareProtectedItem)

	if vault == nil {
		return result, nil
	}

	_, rgName, vaultName, err := ParseVaultId(ptr.Deref(vault.ID, ""))
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

func (c *rwxBackupClient) ListStorageContainers(ctx context.Context, vault *armrecoveryservices.Vault) (map[string]*armrecoveryservicesbackup.AzureStorageContainer, error) {
	result := make(map[string]*armrecoveryservicesbackup.AzureStorageContainer)

	if vault == nil {
		return result, nil
	}

	_, rgName, vaultName, err := ParseVaultId(ptr.Deref(vault.ID, ""))
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

func (c *rwxBackupClient) HasProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error) {
	if vault == nil {
		return false, nil
	}

	_, rgName, vaultName, err := ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return false, err
	}

	protectedItems, err := c.ListProtectedItems(ctx, vaultName, rgName)
	if err != nil {
		return false, nil
	}

	return len(protectedItems) > 0, nil
}

func (c *rwxBackupClient) HasProtectionContainers(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error) {
	if vault == nil {
		return false, nil
	}

	_, rgName, vaultName, err := ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return false, err
	}

	protectedItems, err := c.GetStorageContainers(ctx, rgName, vaultName)
	if err != nil {
		return false, nil
	}

	return len(protectedItems) > 0, nil
}

func (c *rwxBackupClient) DisableSoftDelete(ctx context.Context, vault *armrecoveryservices.Vault) error {
	if vault == nil {
		return nil
	}

	_, rgName, vaultName, err := ParseVaultId(ptr.Deref(vault.ID, ""))
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

func (c *rwxBackupClient) StopFileShareProtection(ctx context.Context, protectedId string, protectedFileShare *armrecoveryservicesbackup.AzureFileshareProtectedItem) error {
	if protectedFileShare == nil {
		return nil
	}

	if *protectedFileShare.ProtectionState == armrecoveryservicesbackup.ProtectionStateProtectionStopped {
		return nil
	}

	protected, err := c.GetProtectedItem(ctx, protectedId)
	if err != nil {
		return err
	}

	protectedFileShare.PolicyID = to.Ptr("")
	protectedFileShare.ProtectionState = to.Ptr(armrecoveryservicesbackup.ProtectionStateProtectionStopped)
	protected.Properties = protectedFileShare

	return c.UpdateProtectedItem(ctx, protected)
}

func (c *rwxBackupClient) RemoveProtection(ctx context.Context, protectedId string) error {

	_, rgName, vaultName, containerName, protectedName, err := ParseProtectedItemId(protectedId)
	if err != nil {
		return err
	}

	//Remove Protection
	fileShareName := GetFileShareName(protectedName)
	err = c.Client.RemoveProtection(ctx, vaultName, rgName, containerName, fileShareName)
	if err != nil {
		return err
	}

	return nil
}

func (c *rwxBackupClient) UnregisterContainer(ctx context.Context, containerId string) error {

	_, rgName, vaultName, containerName, err := ParseContainerId(containerId)
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

func (c *rwxBackupClient) DeleteVault(ctx context.Context, vault *armrecoveryservices.Vault) error {
	if vault == nil {
		return nil
	}

	_, rgName, vaultName, err := ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return err
	}

	err = c.Client.DeleteVault(ctx, rgName, vaultName)
	return err
}
