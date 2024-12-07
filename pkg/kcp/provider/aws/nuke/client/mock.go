package client

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"k8s.io/utils/ptr"
)

type mockClient struct {
	vaultName      string
	recoveryPoints []types.RecoveryPointByBackupVault
}

func NewMock(scopeId string) awsclient.SkrClientProvider[Client] {
	mock := mockClient{}
	mock.init(scopeId)
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		return &mock, nil
	}
}

func (c *mockClient) init(scopeId string) {
	vaultName := fmt.Sprintf("cm-%s", scopeId)
	c.vaultName = vaultName
	c.recoveryPoints = []types.RecoveryPointByBackupVault{
		{
			BackupVaultName:  ptr.To(vaultName),
			RecoveryPointArn: ptr.To("arn:"),
		},
		{
			BackupVaultName:  ptr.To(vaultName),
			RecoveryPointArn: ptr.To("arn:"),
		},
	}
}

func (c *mockClient) ListRecoveryPointsForVault(ctx context.Context, accountId, backupVaultName string) ([]types.RecoveryPointByBackupVault, error) {
	if c.vaultName != backupVaultName {
		return []types.RecoveryPointByBackupVault{}, nil
	}
	return c.recoveryPoints, nil
}

func (c *mockClient) DeleteRecoveryPoint(ctx context.Context, backupVaultName, recoveryPointArn string) (*backup.DeleteRecoveryPointOutput, error) {
	logger := composed.LoggerFromCtx(ctx)
	for i, rPoint := range c.recoveryPoints {
		if ptr.Deref(rPoint.RecoveryPointArn, "") == recoveryPointArn {
			c.recoveryPoints = append(c.recoveryPoints[:i], c.recoveryPoints[i+1:]...)
			logger.WithName("DeleteRecoveryPoint - mock").Info(fmt.Sprintf("Length :: %d", len(c.recoveryPoints)))
		}
	}
	return &backup.DeleteRecoveryPointOutput{}, nil
}
