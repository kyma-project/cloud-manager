package client

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"k8s.io/utils/ptr"
)

type LocalClient interface {
	IsNotFound(err error) bool
}

type Client interface {
	LocalClient
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

func newLocalClient() *localClient {
	return &localClient{}
}

func newClient(svc *backup.Client) Client {
	return &client{
		localClient: *newLocalClient(),
		svc:         svc,
	}
}

type localClient struct {
}

type client struct {
	localClient
	svc *backup.Client
}

func (c *localClient) IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var resourceNotFoundException *backuptypes.ResourceNotFoundException
	return errors.As(err, &resourceNotFoundException)
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
