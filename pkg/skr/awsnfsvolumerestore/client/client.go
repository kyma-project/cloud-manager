package client

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"k8s.io/utils/ptr"
)

type Client interface {
	StartRestoreJob(ctx context.Context, params *StartRestoreJobInput) (*backup.StartRestoreJobOutput, error)
	DescribeRestoreJob(ctx context.Context, restoreJobId string) (*backup.DescribeRestoreJobOutput, error)
	GetRecoveryPointRestoreMetadata(ctx context.Context, accountId, backupVaultName, recoveryPointArn string) (*backup.GetRecoveryPointRestoreMetadataOutput, error)
}

type StartRestoreJobInput struct {
	BackupVaultName  string
	IamRoleArn       string
	IdempotencyToken *string
	RecoveryPointArn *string
	RestoreMetadata  map[string]string
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

func (c *client) StartRestoreJob(ctx context.Context, params *StartRestoreJobInput) (*backup.StartRestoreJobOutput, error) {
	in := &backup.StartRestoreJobInput{
		IamRoleArn:       ptr.To(params.IamRoleArn),
		IdempotencyToken: params.IdempotencyToken,
		RecoveryPointArn: params.RecoveryPointArn,
		ResourceType:     ptr.To("EFS"),
		Metadata:         params.RestoreMetadata,
	}

	out, err := c.svc.StartRestoreJob(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *client) DescribeRestoreJob(ctx context.Context, restoreJobId string) (*backup.DescribeRestoreJobOutput, error) {
	in := &backup.DescribeRestoreJobInput{
		RestoreJobId: ptr.To(restoreJobId),
	}
	out, err := c.svc.DescribeRestoreJob(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *client) GetRecoveryPointRestoreMetadata(ctx context.Context, accountId, backupVaultName, recoveryPointArn string) (*backup.GetRecoveryPointRestoreMetadataOutput, error) {
	in := &backup.GetRecoveryPointRestoreMetadataInput{
		BackupVaultName:      ptr.To(backupVaultName),
		RecoveryPointArn:     ptr.To(recoveryPointArn),
		BackupVaultAccountId: ptr.To(accountId),
	}
	out, err := c.svc.GetRecoveryPointRestoreMetadata(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}
