package client

import (
	"context"
	"errors"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
)

func newVaultMockClient() *vaultClient {
	return &vaultClient{}
}

type vaultMockClient struct {
	vaultClient
}

func (m *vaultMockClient) CreateVault(ctx context.Context, resourceGroupName string, vaultName string, location string) (*string, error) {

	// unhappy path
	if resourceGroupName == "http-error" {
		return nil, errors.New("some error")
	}

	// happy path
	id := "id"
	return &id, nil

}

func (m *vaultMockClient) DeleteVault(ctx context.Context, resourceGroupName string, vaultName string) error {

	// happy path
	return nil

}

func (m *vaultMockClient) ListVaults(ctx context.Context) ([]*armrecoveryservices.Vault, error) {

	var res []*armrecoveryservices.Vault

	// happy path
	return res, nil

}
