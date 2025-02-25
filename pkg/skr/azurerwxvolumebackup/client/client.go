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
}

type client struct {
	VaultClient
	BackupClient
	ProtectionPoliciesClient
	RecoveryPointClient
}

func NewClientProvider() azureclient.ClientProvider[Client] {

	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string) (Client, error) {
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

		c = client{
			vc,
			bc,
			ppc,
			rpc,
		}

		return c, nil

	}

}
