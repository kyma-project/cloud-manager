package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"slices"
)

type VaultClient interface {
	CreateVault(ctx context.Context, resourceGroupName string, vaultName string, location string) error
	DeleteVault(ctx context.Context, resourceGroupName string, vaultName string) error
	ListVaults(ctx context.Context) ([]*armrecoveryservices.Vault, error)
}

type vaultClient struct {
	azureClient *armrecoveryservices.VaultsClient
}

func NewVaultClient(subscriptionId string, cred *azidentity.ClientSecretCredential) (VaultClient, error) {

	vc, err := armrecoveryservices.NewVaultsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}

	return vaultClient{vc}, nil
}

func (c vaultClient) vaultExists(ctx context.Context, location string) (bool, error) {
	vaults, err := c.ListVaults(ctx)
	if err != nil {
		return false, err
	}

	slices.ContainsFunc(vaults, func(vault *armrecoveryservices.Vault) bool {

		if vault.Location == nil || vault.Tags == nil {
			return false
		}

		_, tagExists := vault.Tags["cloud-manager"]

		return *vault.Location == location && tagExists

	})
	return false, nil
}

func (c vaultClient) CreateVault(ctx context.Context, resourceGroupName string, vaultName string, location string) error {

	// Fail if vault exists
	exists, err := c.vaultExists(ctx, location)
	if err != nil {
		return err
	}
	if exists {
		return errors.New(fmt.Sprintf("Vault already exists in %s", location))
	}

	_, err = c.azureClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		vaultName,
		armrecoveryservices.Vault{
			Location: to.Ptr(location),
			Properties: to.Ptr(armrecoveryservices.VaultProperties{
				PublicNetworkAccess: to.Ptr(armrecoveryservices.PublicNetworkAccessEnabled),
			}),
			SKU: to.Ptr(armrecoveryservices.SKU{
				Name: to.Ptr(armrecoveryservices.SKUNameStandard),
			}),
			Tags: map[string]*string{"cloud-manager": to.Ptr("rwxVolumeBackup")},
		},
		nil,
	)

	if err != nil {
		return err
	}

	return nil

}

func (c vaultClient) DeleteVault(ctx context.Context, resourceGroupName string, vaultName string) error {

	_, err := c.azureClient.Delete(
		ctx,
		resourceGroupName,
		vaultName,
		to.Ptr(armrecoveryservices.VaultsClientDeleteOptions{}),
	)

	if err != nil {
		return err
	}

	return nil
}

func (c vaultClient) ListVaults(ctx context.Context) ([]*armrecoveryservices.Vault, error) {

	pager := c.azureClient.NewListBySubscriptionIDPager(
		&armrecoveryservices.VaultsClientListBySubscriptionIDOptions{},
	)

	var vaults []*armrecoveryservices.Vault
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return vaults, err
		}
		for _, v := range page.VaultList.Value {
			vaults = append(vaults, v)
		}
	}
	return vaults, nil

}
