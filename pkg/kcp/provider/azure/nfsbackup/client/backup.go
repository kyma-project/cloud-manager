package client

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservicesbackup/v4"
)

type BackupClient interface {
	TriggerBackup(ctx context.Context)
}

type backupClient struct {
	*armrecoveryservicesbackup.BackupsClient
}

func NewBackupClient(subscriptionId string, cred *azidentity.ClientSecretCredential) (BackupClient, error) {

	bc, err := armrecoveryservicesbackup.NewBackupsClient(subscriptionId, cred, nil)

	if err != nil {
		return nil, err
	}

	return backupClient{bc}, nil
}

func (c backupClient) TriggerBackup(ctx context.Context) {
	// TODO: trigger logic here
}
