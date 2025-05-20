package client

import (
	"context"

	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
)

const (
	AzureRecoveryVault       = "AzureRecoveryVault"
	AzureStorageContainer    = "AzureStorageContainer"
	AzureFileShareProtection = "AzureFileShareProtection"
)

type NukeRwxBackupClient interface {
	azurerwxvolumebackupclient.RwxBackupClient
}

type nukeRwxBackupClient struct {
	azurerwxvolumebackupclient.RwxBackupClient
}

func NewClientProvider() azureclient.ClientProvider[NukeRwxBackupClient] {
	return NukeProvider(azurerwxvolumebackupclient.NewClientProvider())
}

func NukeProvider(backupClientProvider azureclient.ClientProvider[azurerwxvolumebackupclient.Client]) azureclient.ClientProvider[NukeRwxBackupClient] {

	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (NukeRwxBackupClient, error) {
		rwxBackupProvider := azurerwxvolumebackupclient.RwxBackupClientProvider(backupClientProvider)
		client, err := rwxBackupProvider(ctx, clientId, clientSecret, subscriptionId, tenantId, auxiliaryTenants...)

		if err != nil {
			return nil, err
		}

		return &nukeRwxBackupClient{client}, nil
	}
}
