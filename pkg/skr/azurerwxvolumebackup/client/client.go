package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

type Client interface {
	VaultClient
	BackupClient
	ProtectionPoliciesClient
	RecoveryPointClient
	JobsClient
	RestoreClient
	BackupProtectableItemsClient
	ProtectedItemsClient
	BackupProtectedItemsClient
	VaultConfigClient
}

type client struct {
	VaultClient
	BackupClient
	ProtectionPoliciesClient
	RecoveryPointClient
	JobsClient
	RestoreClient
	BackupProtectableItemsClient
	ProtectedItemsClient
	BackupProtectedItemsClient
	VaultConfigClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {

	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {
		var c Client
		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, azureclient.NewCredentialOptionsBuilder().Build())
		if err != nil {
			return nil, err
		}

		recoveryServicesFactory, err := armrecoveryservices.NewClientFactory(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())

		if err != nil {
			return nil, err
		}

		recoveryServicesBackupFactory, err := armrecoveryservicesbackup.NewClientFactory(subscriptionId, cred, azureclient.NewClientOptionsBuilder().Build())

		if err != nil {
			return nil, err
		}

		c = client{
			NewVaultClient(recoveryServicesFactory.NewVaultsClient()),
			NewBackupClient(recoveryServicesBackupFactory.NewBackupsClient()),
			NewProtectionPoliciesClient(recoveryServicesBackupFactory.NewProtectionPoliciesClient()),
			NewRecoveryPointClient(recoveryServicesBackupFactory.NewRecoveryPointsClient()),
			NewJobsClient(recoveryServicesBackupFactory.NewBackupJobsClient(), recoveryServicesBackupFactory.NewJobDetailsClient()),
			NewRestoreClient(recoveryServicesBackupFactory.NewRestoresClient()),
			NewBackupProtectableItemsClient(recoveryServicesBackupFactory.NewBackupProtectableItemsClient()),
			NewProtectedItemsClient(recoveryServicesBackupFactory.NewProtectedItemsClient()),
			NewBackupProtectedItemsClient(recoveryServicesBackupFactory.NewBackupProtectedItemsClient()),
			NewVaultConfigClient(
				recoveryServicesBackupFactory.NewBackupResourceVaultConfigsClient(),
				recoveryServicesBackupFactory.NewBackupProtectionContainersClient(),
				recoveryServicesBackupFactory.NewProtectionContainersClient(),
			),
		}

		return c, nil
	}
}
