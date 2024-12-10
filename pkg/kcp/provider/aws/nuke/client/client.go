package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/aws/aws-sdk-go-v2/service/backup/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"k8s.io/utils/ptr"
)

type Client interface {
	ListRecoveryPointsForVault(ctx context.Context, accountId, backupVaultName string) ([]types.RecoveryPointByBackupVault, error)
	DeleteRecoveryPoint(ctx context.Context, backupVaultName, recoveryPointArn string) (*backup.DeleteRecoveryPointOutput, error)
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(backup.NewFromConfig(cfg)), nil
	}
}

func newClient(svc *backup.Client) Client {
	return &client{
		svc: svc,
	}
}

type client struct {
	svc *backup.Client
}

func (c *client) ListRecoveryPointsForVault(ctx context.Context, accountId, backupVaultName string) ([]types.RecoveryPointByBackupVault, error) {
	in := &backup.ListRecoveryPointsByBackupVaultInput{
		BackupVaultName:      ptr.To(backupVaultName),
		BackupVaultAccountId: ptr.To(accountId),
	}
	out, err := c.svc.ListRecoveryPointsByBackupVault(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.RecoveryPoints, nil
}

func (c *client) DeleteRecoveryPoint(ctx context.Context, backupVaultName, recoveryPointArn string) (*backup.DeleteRecoveryPointOutput, error) {
	in := &backup.DeleteRecoveryPointInput{
		BackupVaultName:  ptr.To(backupVaultName),
		RecoveryPointArn: ptr.To(recoveryPointArn),
	}
	out, err := c.svc.DeleteRecoveryPoint(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}
