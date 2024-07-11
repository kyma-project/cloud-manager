package client

import (
	"context"
	"errors"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	"k8s.io/utils/pointer"
)

type Client interface {
	IsNotFound(err error) bool
	IsAlreadyExists(err error) bool
	ListTags(ctx context.Context, resourceArn string) (map[string]string, error)
	//ListBackupVaults(ctx context.Context) ([]backuptypes.BackupVaultListMember, error)
	DescribeBackupVault(ctx context.Context, backupVaultName string) (*backup.DescribeBackupVaultOutput, error)
	CreateBackupVault(ctx context.Context, name string, tags map[string]string) (string, error)
	DeleteBackupVault(ctx context.Context, name string) error
	StartBackupJob(ctx context.Context, params *StartBackupJobInput) (*backup.StartBackupJobOutput, error)
	ListBackupJobs(ctx context.Context, in *backup.ListBackupJobsInput) ([]backuptypes.BackupJob, error)
	DescribeBackupJob(ctx context.Context, backupJobId string) (*backup.DescribeBackupJobOutput, error)
}

type StartBackupJobInput struct {
	BackupValueName            string
	ResourceArn                string
	IamRoleArn                 string
	IdempotencyToken           *string
	DeleteAfterDays            *int64
	MoveToColdStorageAfterDays *int64
	RecoveryPointTags          map[string]string
}

func NewClientProvider() awsclient.SkrClientProvider[Client] {
	return func(ctx context.Context, region, key, secret, role string) (Client, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return newClient(backup.NewFromConfig(cfg)), nil
	}
}

func newClient(svc *backup.Client) Client {
	return &client{svc: svc}
}

type client struct {
	svc *backup.Client
}

func (c *client) IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	var resourceNotFoundException *backuptypes.ResourceNotFoundException
	return errors.As(err, &resourceNotFoundException)
}

func (c *client) IsAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	var alreadyExistsException *backuptypes.AlreadyExistsException
	return errors.As(err, &alreadyExistsException)
}

func (c *client) ListTags(ctx context.Context, resourceArn string) (map[string]string, error) {
	in := &backup.ListTagsInput{
		ResourceArn: pointer.String(resourceArn),
	}
	out, err := c.svc.ListTags(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.Tags, nil
}

func (c *client) ListBackupVaults(ctx context.Context) ([]backuptypes.BackupVaultListMember, error) {
	in := &backup.ListBackupVaultsInput{}
	out, err := c.svc.ListBackupVaults(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.BackupVaultList, nil
}

func (c *client) DescribeBackupVault(ctx context.Context, backupVaultName string) (*backup.DescribeBackupVaultOutput, error) {
	in := &backup.DescribeBackupVaultInput{
		BackupVaultName: pointer.String(backupVaultName),
	}
	out, err := c.svc.DescribeBackupVault(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *client) CreateBackupVault(ctx context.Context, name string, tags map[string]string) (string, error) {
	in := &backup.CreateBackupVaultInput{
		BackupVaultName:  pointer.String(name),
		BackupVaultTags:  tags,
		CreatorRequestId: pointer.String(name),
	}
	out, err := c.svc.CreateBackupVault(ctx, in)
	if err != nil {
		return "", err
	}
	return pointer.StringDeref(out.BackupVaultArn, ""), nil
}

func (c *client) DeleteBackupVault(ctx context.Context, name string) error {
	in := &backup.DeleteBackupVaultInput{
		BackupVaultName: pointer.String(name),
	}
	_, err := c.svc.DeleteBackupVault(ctx, in)
	return err
}

func (c *client) StartBackupJob(ctx context.Context, params *StartBackupJobInput) (*backup.StartBackupJobOutput, error) {
	var lifecycle *backuptypes.Lifecycle
	if params.DeleteAfterDays != nil || params.MoveToColdStorageAfterDays != nil {
		lifecycle = &backuptypes.Lifecycle{
			DeleteAfterDays:            params.DeleteAfterDays,
			MoveToColdStorageAfterDays: params.MoveToColdStorageAfterDays,
		}
	}
	in := &backup.StartBackupJobInput{
		BackupVaultName:   pointer.String(params.BackupValueName),
		IamRoleArn:        pointer.String(params.IamRoleArn),
		ResourceArn:       pointer.String(params.ResourceArn),
		BackupOptions:     nil, // ???
		IdempotencyToken:  params.IdempotencyToken,
		Lifecycle:         lifecycle,
		RecoveryPointTags: params.RecoveryPointTags,
	}
	out, err := c.svc.StartBackupJob(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *client) ListBackupJobs(ctx context.Context, in *backup.ListBackupJobsInput) ([]backuptypes.BackupJob, error) {
	out, err := c.svc.ListBackupJobs(ctx, in)
	if err != nil {
		return nil, err
	}
	return out.BackupJobs, nil
}

func (c *client) DescribeBackupJob(ctx context.Context, backupJobId string) (*backup.DescribeBackupJobOutput, error) {
	in := &backup.DescribeBackupJobInput{
		BackupJobId: pointer.String(backupJobId),
	}
	out, err := c.svc.DescribeBackupJob(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}
