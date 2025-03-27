package client

import (
	"context"
	"errors"
)

func newBackupMockClient() *backupClient {
	return &backupClient{}
}

type backupMockClient struct {
	backupClient
}

func (m *backupMockClient) TriggerBackup(ctx context.Context, vaultName, resourceGroupName, containerName, protectedItemName, location string) error {

	// unhappy path
	if vaultName == "exactly 1 - fail" {
		return errors.New("failing test")
	}

	// happy path
	return nil
}
