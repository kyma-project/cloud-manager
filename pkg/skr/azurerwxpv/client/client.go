package client

import (
	"context"

	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
)

type Client interface {
	azurerwxvolumebackupclient.RwxBackupClient
	azurerwxvolumebackupclient.FileShareClient
}

type client struct {
	azurerwxvolumebackupclient.RwxBackupClient
	azurerwxvolumebackupclient.FileShareClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {

	backupProvider := azurerwxvolumebackupclient.NewClientProvider()
	rwxBackupProvider := azurerwxvolumebackupclient.RwxBackupClientProvider(backupProvider)
	fscProvider := azurerwxvolumebackupclient.FileShareClientProvider(backupProvider)

	return NewAzurePvProvider(rwxBackupProvider, fscProvider)
}

func NewAzurePvProvider(rwxBackupProvider azureclient.ClientProvider[azurerwxvolumebackupclient.RwxBackupClient], fscProvider azureclient.ClientProvider[azurerwxvolumebackupclient.FileShareClient]) azureclient.ClientProvider[Client] {

	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {
		rwxBackupClient, err := rwxBackupProvider(ctx, clientId, clientSecret, subscriptionId, tenantId, auxiliaryTenants...)
		if err != nil {
			return nil, err
		}

		fsc, err := fscProvider(ctx, clientId, clientSecret, subscriptionId, tenantId, auxiliaryTenants...)
		if err != nil {
			return nil, err
		}

		return client{rwxBackupClient, fsc}, nil
	}
}
