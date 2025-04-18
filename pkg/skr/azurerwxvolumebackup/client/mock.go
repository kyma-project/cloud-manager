package client

import (
	"context"

	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
)

func NewMockClient() azureclient.ClientProvider[Client] {
	return func(ctx context.Context, _, _, _, _ string, _ ...string) (Client, error) {
		jobsMock = &jobsMockClient{
			jobsClient: *newJobsMockClient(),
		}
		restoreMock = &restoreMockClient{
			restoreClient: *newRestoreMockClient(),
		}

		// TODO: is this okay initialization?
		vaultMock := &vaultMockClient{vaultClient: *newVaultMockClient()}

		backupMock := &backupMockClient{backupClient: *newBackupMockClient()}
		backupProtectedItemsMock := &backupProtectedItemsMockClient{backupProtectedItemsClient: *newBackupProtectedItemsMockClient()}
		protectionPoliciesMock := &protectionPoliciesMockClient{protectionPoliciesClient: *newProtectionPoliciesMockClient()}
		backupProtectableItemsMock := &backupProtectableItemsMockClient{backupProtectableItemsClient: *newBackupProtectableItemsMockClient()}
		protectedItemsMock := &protectedItemsMockClient{protectedItemsClient: *newProtectedItemsMockClient()}

		return client{
			vaultMock,
			backupMock,
			protectionPoliciesMock,
			nil,
			jobsMock,
			restoreMock,
			backupProtectableItemsMock,
			protectedItemsMock,
			backupProtectedItemsMock,
			newVaultConfigMockClient(),
		}, nil
	}
}
