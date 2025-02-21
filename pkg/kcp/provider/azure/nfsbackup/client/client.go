package client

import "github.com/Azure/azure-sdk-for-go/sdk/azidentity"

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

func NewClient(subscriptionId string) (Client, error) {
	var c Client
	cred, err := azidentity.NewDefaultAzureCredential(nil)
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
