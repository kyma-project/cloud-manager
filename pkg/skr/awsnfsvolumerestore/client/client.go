package client

import (
	"context"
	"errors"
	"time"

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
	ListRestoreJobsByRecoveryPoint(ctx context.Context, recoveryPointArn string) ([]backuptypes.RestoreJobsListMember, error)
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

func (c *client) ListRestoreJobsByRecoveryPoint(ctx context.Context, recoveryPointArn string) ([]backuptypes.RestoreJobsListMember, error) {
	// AWS doesn't support filtering by recovery point ARN, so we list recent EFS restore jobs
	// and filter manually
	createdAfter := time.Now().Add(-24 * time.Hour) // Look back 24 hours
	in := &backup.ListRestoreJobsInput{
		ByResourceType: ptr.To("EFS"),
		ByCreatedAfter: ptr.To(createdAfter),
	}
	out, err := c.svc.ListRestoreJobs(ctx, in)
	if err != nil {
		return nil, err
	}

	// Filter by recovery point ARN
	var filtered []backuptypes.RestoreJobsListMember
	for _, job := range out.RestoreJobs {
		// Need to describe each job to get the recovery point ARN
		// This is inefficient but AWS API doesn't provide a better way
		details, err := c.DescribeRestoreJob(ctx, ptr.Deref(job.RestoreJobId, ""))
		if err != nil {
			continue // Skip jobs we can't describe
		}
		if ptr.Deref(details.RecoveryPointArn, "") == recoveryPointArn {
			filtered = append(filtered, job)
		}
	}

	return filtered, nil
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
