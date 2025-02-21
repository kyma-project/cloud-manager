package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
)

type VaultClient interface {
	CreateVault(ctx context.Context)
	DeleteVault(ctx context.Context)
}

type vaultClient struct {
	*armrecoveryservices.VaultsClient
}

func NewVaultClient(subscriptionId string, cred *azidentity.DefaultAzureCredential) (VaultClient, error) {

	vc, err := armrecoveryservices.NewVaultsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}

	return vaultClient{vc}, nil
}

func (c vaultClient) CreateVault(ctx context.Context) {
	// TODO: implementation details
}

func (c vaultClient) DeleteVault(ctx context.Context) {
	// TODO: implementation details
}
