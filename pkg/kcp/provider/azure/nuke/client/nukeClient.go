package client

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"k8s.io/utils/ptr"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
)

type NukeRwxBackupClient interface {
	ListRwxVolumeBackupVaults(ctx context.Context) ([]*armrecoveryservices.Vault, error)
	ListFileShareProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) ([]*armrecoveryservicesbackup.ProtectedItemResource, error)
	RemoveProtection(ctx context.Context, protected *armrecoveryservicesbackup.ProtectedItemResource) error
	HasProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) (bool, error)
	DisableSoftDelete(ctx context.Context, vault *armrecoveryservices.Vault) error
	DeleteVault(ctx context.Context, vault *armrecoveryservices.Vault, containers []string) error
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

func (c *nukeRwxBackupClient) ListFileShareProtectedItems(ctx context.Context, vault *armrecoveryservices.Vault) ([]*armrecoveryservicesbackup.ProtectedItemResource, error) {
	var result []*armrecoveryservicesbackup.ProtectedItemResource

	if vault == nil {
		return result, nil
	}

	logger := composed.LoggerFromCtx(ctx)

	_, rgName, vaultName, err := azurerwxvolumebackupclient.ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return result, nil
	}

	protectedItems, err := c.ListProtectedItems(ctx, vaultName, rgName)
	if err != nil {
		return result, nil
	}

	logger.Info(fmt.Sprintf("Number of Protected Items : %d", len(protectedItems)))
	for _, item := range protectedItems {

		switch protected := item.Properties.(type) {
		case *armrecoveryservicesbackup.AzureFileshareProtectedItem:
			if ptr.Deref(protected.ProtectionState, "") == armrecoveryservicesbackup.ProtectionStateProtected {
				result = append(result, item)
			}
		default:
			continue
		}
	}

	return result, nil

}

func (c *nukeRwxBackupClient) RemoveProtection(ctx context.Context, protected *armrecoveryservicesbackup.ProtectedItemResource) error {

	if protected == nil {
		return nil
	}

	_, rgName, vaultName, containerName, protectedName, err := azurerwxvolumebackupclient.ParseProtectedItemId(*protected.ID)
	if err != nil {
		return err
	}

	return c.Client.RemoveProtection(ctx, vaultName, rgName, containerName, protectedName)
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

	composed.LoggerFromCtx(ctx).Info(fmt.Sprintf("Protected Item Count : %d", len(protectedItems)))
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

func (c *nukeRwxBackupClient) DeleteVault(ctx context.Context, vault *armrecoveryservices.Vault, containers []string) error {
	if vault == nil {
		return nil
	}

	_, rgName, vaultName, err := azurerwxvolumebackupclient.ParseVaultId(ptr.Deref(vault.ID, ""))
	if err != nil {
		return err
	}

	//Unregister the containers
	for _, containerName := range containers {
		err = c.UnregisterContainer(ctx, rgName, vaultName, containerName)
		if err != nil {
			return err
		}
	}

	return c.Client.DeleteVault(ctx, rgName, vaultName)
}
