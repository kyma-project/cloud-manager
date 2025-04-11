package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
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
}

func NewClientProvider() azureclient.ClientProvider[Client] {

	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, auxiliaryTenants ...string) (Client, error) {
		var c Client
		cred, err := azidentity.NewClientSecretCredential(tenantId, clientId, clientSecret, &azidentity.ClientSecretCredentialOptions{})
		if err != nil {
			return nil, err
		}

		vc, err := NewVaultClient(subscriptionId, cred)
		if err != nil {
			return nil, err
		}

		bc, err := NewBackupClient(subscriptionId, cred)
		if err != nil {
			return nil, err
		}

		ppc, err := NewProtectionPoliciesClient(subscriptionId, cred)
		if err != nil {
			return nil, err
		}

		rpc, err := NewRecoveryPointClient(subscriptionId, cred)
		if err != nil {
			return nil, err
		}

		jc, err := NewJobsClient(subscriptionId, cred)
		if err != nil {
			return nil, err
		}

		rc, err := NewRestoreClient(subscriptionId, cred)
		if err != nil {
			return nil, err
		}

		bpic, err := NewBackupProtectableItemsClient(subscriptionId, cred)
		if err != nil {
			return nil, err
		}

		pic, err := NewProtectedItemsClient(subscriptionId, cred)
		if err != nil {
			return nil, err
		}

		bprotectedic, err := NewBackupProtectedItemsClient(subscriptionId, cred)
		if err != nil {
			return nil, err
		}

		c = client{
			vc,
			bc,
			ppc,
			rpc,
			jc,
			rc,
			bpic,
			pic,
			bprotectedic,
		}

		return c, nil
	}
}
