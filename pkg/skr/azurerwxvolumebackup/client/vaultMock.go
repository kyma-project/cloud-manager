package client

import (
	"context"
	"errors"
	"fmt"
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
	if ctx.Value("CreateVault") == "fail" {
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

	location := "uswest"
	value := "value"
	vaultName := fmt.Sprintf("cm-vault-%s", location)
	res := []*armrecoveryservices.Vault{
		&armrecoveryservices.Vault{
			Location:   &location,
			Etag:       nil,
			Identity:   nil,
			Properties: nil,
			SKU:        nil,
			Tags:       map[string]*string{"cloud-manager": &value},
			ID:         nil,
			Name:       &vaultName,
			SystemData: nil,
			Type:       nil,
		},
	}

	// happy path
	return res, nil

}
