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
	if ctx.Value("TriggerBackup") == "fail" {
		return errors.New("failing test")
	}

	// happy path
	return nil
}
