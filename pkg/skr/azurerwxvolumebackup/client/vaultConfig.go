package client

import (
	"context"
	"errors"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

type VaultConfigClient interface {
	GetVaultConfig(ctx context.Context, resourceGroupName, vaultName string) (*armrecoveryservicesbackup.BackupResourceVaultConfigResource, error)
	PutVaultConfig(ctx context.Context, resourceGroupName, vaultName string, config *armrecoveryservicesbackup.BackupResourceVaultConfigResource) error
	UnregisterContainer(ctx context.Context, resourceGroupName, vaultName, containerName string) error
}

type vaultConfigClient struct {
	configClient       *armrecoveryservicesbackup.BackupResourceVaultConfigsClient
	registrationClient *armrecoveryservices.RegisteredIdentitiesClient
}

func NewVaultConfigClient(subscriptionId string, cred *azidentity.ClientSecretCredential) (VaultConfigClient, error) {

	vc, err := armrecoveryservicesbackup.NewBackupResourceVaultConfigsClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}

	rc, err := armrecoveryservices.NewRegisteredIdentitiesClient(subscriptionId, cred, nil)
	if err != nil {
		return nil, err
	}
	return vaultConfigClient{configClient: vc, registrationClient: rc}, nil
}

func (c vaultConfigClient) GetVaultConfig(ctx context.Context, resourceGroupName, vaultName string) (*armrecoveryservicesbackup.BackupResourceVaultConfigResource, error) {

	result, err := c.configClient.Get(
		ctx,
		resourceGroupName,
		vaultName, nil)

	if err != nil {
		log.Println("failed to get vault config: " + err.Error())
		return nil, err
	}

	return &result.BackupResourceVaultConfigResource, nil

}

func (c vaultConfigClient) PutVaultConfig(ctx context.Context, resourceGroupName, vaultName string, config *armrecoveryservicesbackup.BackupResourceVaultConfigResource) error {

	if config == nil {
		return errors.New("vault config cannot be nil")
	}

	_, err := c.configClient.Put(
		ctx,
		resourceGroupName,
		vaultName,
		*config,
		nil)

	return err
}

func (c vaultConfigClient) UnregisterContainer(ctx context.Context, resourceGroupName, vaultName, containerName string) error {
	_, err := c.registrationClient.Delete(
		ctx,
		resourceGroupName,
		vaultName,
		containerName,
		nil)

	return err
}
