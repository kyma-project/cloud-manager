package client

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"k8s.io/utils/ptr"

	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
)

type NukeRwxBackupClient interface {
	ListRwxVolumeBackupVaults(ctx context.Context) ([]*armrecoveryservices.Vault, error)
	ListFileShareProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) ([]*armrecoveryservicesbackup.ProtectedItemResource, error)
	RemoveProtection(ctx context.Context, protected *armrecoveryservicesbackup.ProtectedItemResource) error
	HasProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error)
	DeleteVault(ctx context.Context, vault *armrecoveryservices.Vault) error
}

type nukeRwxBackupClient struct {
	azurebackupclient.Client
}

func NewClientProvider() azureclient.ClientProvider[NukeRwxBackupClient] {
	return NukeProvider(azurebackupclient.NewClientProvider())
}

func NukeProvider(backupProvider azureclient.ClientProvider[azurebackupclient.Client]) azureclient.ClientProvider[NukeRwxBackupClient] {

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
		if value, exists := vault.Tags[tagNameCloudManager]; exists && ptr.Deref(value, "") == tagValueRwxVolumeBackup {
			result = append(result, vault)
		}
	}

	return result, nil
}

func (c *nukeRwxBackupClient) ListFileShareProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) ([]*armrecoveryservicesbackup.ProtectedItemResource, error) {
	var result []*armrecoveryservicesbackup.ProtectedItemResource

	if vault == nil {
		return result, nil
	}

	_, rgName, vaultName, err := ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return result, nil
	}

	protectedItems, err := c.Client.ListProtectedItems(ctx, vaultName, rgName)
	if err != nil {
		return result, nil
	}

	for _, item := range protectedItems {

		fileShare, okay := item.Properties.(*armrecoveryservicesbackup.AzureFileshareProtectedItem)
		if !okay {
			continue
		}
		if ptr.Deref(fileShare.ProtectionState, "") != armrecoveryservicesbackup.ProtectionStateProtected {
			continue
		}
		result = append(result, item)
	}

	return result, nil

}

func (c *nukeRwxBackupClient) RemoveProtection(ctx context.Context, protected *armrecoveryservicesbackup.ProtectedItemResource) error {

	if protected == nil {
		return nil
	}

	_, rgName, vaultName, containerName, protectedName, err := ParseProtectedItemId(*protected.ID)
	if err != nil {
		return err
	}

	return c.Client.RemoveProtection(ctx, vaultName, rgName, containerName, protectedName)
}

func (c *nukeRwxBackupClient) HasProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error) {
	if vault == nil {
		return false, nil
	}

	_, rgName, vaultName, err := ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return false, err
	}

	protectedItems, err := c.Client.ListProtectedItems(ctx, vaultName, rgName)
	if err != nil {
		return false, nil
	}

	return len(protectedItems) > 0, nil
}

func (c *nukeRwxBackupClient) DeleteVault(ctx context.Context, vault *armrecoveryservices.Vault) error {
	if vault == nil {
		return nil
	}

	_, rgName, vaultName, err := ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return err
	}

	return c.Client.DeleteVault(ctx, rgName, vaultName)
}
