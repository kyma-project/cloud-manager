package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
	"time"
)

type BackupClient interface {
	TriggerBackup(ctx context.Context, vaultName, resourceGroupName, containerName, protectedItemName, location string) error
}

type backupClient struct {
	azureClient *armrecoveryservicesbackup.BackupsClient
}

func NewBackupClient(subscriptionId string, cred *azidentity.ClientSecretCredential) (BackupClient, error) {

	bc, err := armrecoveryservicesbackup.NewBackupsClient(subscriptionId, cred, nil)

	if err != nil {
		return nil, err
	}

	return backupClient{bc}, nil
}

func (c backupClient) TriggerBackup(ctx context.Context, vaultName, resourceGroupName, containerName, protectedItemName, location string) error {

	// Long deletion time to help separate user defined values
	expiryTime := time.Date(2124, time.January, 1, 0, 0, 0, 0, time.UTC)
	params := armrecoveryservicesbackup.BackupRequestResource{
		ETag:     nil,
		Location: to.Ptr(location),
		Properties: to.Ptr(armrecoveryservicesbackup.AzureFileShareBackupRequest{
			ObjectType:                   to.Ptr("AzureFileShareBackupRequest"),
			RecoveryPointExpiryTimeInUTC: to.Ptr(expiryTime),
		}),
		Tags: nil,
	}
	_, err := c.azureClient.Trigger(ctx, vaultName, resourceGroupName, "Azure", containerName, protectedItemName, params, nil)
	if err != nil {
		return err
	}
	return nil
}
