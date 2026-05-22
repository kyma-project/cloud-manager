package client

import (
	"context"

	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awscertificateclient "github.com/kyma-project/cloud-manager/pkg/skr/awscertificate/client"
	awsnfsvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	awswebaclclient "github.com/kyma-project/cloud-manager/pkg/skr/awswebacl/client"
)

type NukeClient interface {
	awsnfsvolumebackupclient.Client

	// WebACL methods
	ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error)
	GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error)
	DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error
	ListTagsForWebACL(ctx context.Context, resourceArn string) ([]wafv2types.Tag, error)

	// Certificate methods
	ListCertificates(ctx context.Context) ([]acmtypes.CertificateSummary, error)
	ListCertificateTags(ctx context.Context, arn string) ([]acmtypes.Tag, error)
	DeleteCertificate(ctx context.Context, arn string) error
}

func NewClientProvider() awsclient.SkrClientProvider[NukeClient] {
	return NukeProvider(
		awsnfsvolumebackupclient.NewClientProvider(),
		awswebaclclient.NewClientProvider(),
		awscertificateclient.NewClientProvider(),
	)
}

func NukeProvider(
	backupProvider awsclient.SkrClientProvider[awsnfsvolumebackupclient.Client],
	webAclProvider awsclient.SkrClientProvider[awswebaclclient.Client],
	certificateProvider awsclient.SkrClientProvider[awscertificateclient.Client],
) awsclient.SkrClientProvider[NukeClient] {
	return func(ctx context.Context, account, region, key, secret, role string) (NukeClient, error) {
		backup, err := backupProvider(ctx, account, region, key, secret, role)
		if err != nil {
			return nil, err
		}

		webAcl, err := webAclProvider(ctx, account, region, key, secret, role)
		if err != nil {
			return nil, err
		}

		cert, err := certificateProvider(ctx, account, region, key, secret, role)
		if err != nil {
			return nil, err
		}

		return &client{
			backupClient:      backup,
			webAclClient:      webAcl,
			certificateClient: cert,
		}, nil
	}
}

type client struct {
	backupClient      awsnfsvolumebackupclient.Client
	webAclClient      awswebaclclient.Client
	certificateClient awscertificateclient.Client
}

// Embed backup client methods
func (c *client) IsNotFound(err error) bool {
	return c.backupClient.IsNotFound(err)
}

func (c *client) IsAlreadyExists(err error) bool {
	return c.backupClient.IsAlreadyExists(err)
}

func (c *client) ListTags(ctx context.Context, resourceArn string) (map[string]string, error) {
	return c.backupClient.ListTags(ctx, resourceArn)
}

func (c *client) ListBackupVaults(ctx context.Context) ([]backuptypes.BackupVaultListMember, error) {
	return c.backupClient.ListBackupVaults(ctx)
}

func (c *client) DescribeBackupVault(ctx context.Context, backupVaultName string) (*backup.DescribeBackupVaultOutput, error) {
	return c.backupClient.DescribeBackupVault(ctx, backupVaultName)
}

func (c *client) CreateBackupVault(ctx context.Context, name string, tags map[string]string) (string, error) {
	return c.backupClient.CreateBackupVault(ctx, name, tags)
}

func (c *client) DeleteBackupVault(ctx context.Context, name string) error {
	return c.backupClient.DeleteBackupVault(ctx, name)
}

func (c *client) StartBackupJob(ctx context.Context, params *awsnfsvolumebackupclient.StartBackupJobInput) (*backup.StartBackupJobOutput, error) {
	return c.backupClient.StartBackupJob(ctx, params)
}

func (c *client) DescribeBackupJob(ctx context.Context, backupJobId string) (*backup.DescribeBackupJobOutput, error) {
	return c.backupClient.DescribeBackupJob(ctx, backupJobId)
}

func (c *client) ListRecoveryPointsForVault(ctx context.Context, accountId, backupVaultName string) ([]backuptypes.RecoveryPointByBackupVault, error) {
	return c.backupClient.ListRecoveryPointsForVault(ctx, accountId, backupVaultName)
}

func (c *client) DescribeRecoveryPoint(ctx context.Context, accountId, backupVaultName, recoveryPointArn string) (*backup.DescribeRecoveryPointOutput, error) {
	return c.backupClient.DescribeRecoveryPoint(ctx, accountId, backupVaultName, recoveryPointArn)
}

func (c *client) DeleteRecoveryPoint(ctx context.Context, backupVaultName, recoveryPointArn string) (*backup.DeleteRecoveryPointOutput, error) {
	return c.backupClient.DeleteRecoveryPoint(ctx, backupVaultName, recoveryPointArn)
}

func (c *client) StartCopyJob(ctx context.Context, params *awsnfsvolumebackupclient.StartCopyJobInput) (*backup.StartCopyJobOutput, error) {
	return c.backupClient.StartCopyJob(ctx, params)
}

func (c *client) DescribeCopyJob(ctx context.Context, copyJobId string) (*backup.DescribeCopyJobOutput, error) {
	return c.backupClient.DescribeCopyJob(ctx, copyJobId)
}

// WebACL methods
func (c *client) ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error) {
	return c.webAclClient.ListWebACLs(ctx, scope)
}

func (c *client) GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error) {
	return c.webAclClient.GetWebACL(ctx, name, id, scope)
}

func (c *client) DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error {
	return c.webAclClient.DeleteWebACL(ctx, name, id, scope, lockToken)
}

func (c *client) ListTagsForWebACL(ctx context.Context, resourceArn string) ([]wafv2types.Tag, error) {
	return c.webAclClient.ListTagsForWebACL(ctx, resourceArn)
}

// Certificate methods
func (c *client) ListCertificates(ctx context.Context) ([]acmtypes.CertificateSummary, error) {
	return c.certificateClient.ListCertificates(ctx)
}

func (c *client) ListCertificateTags(ctx context.Context, arn string) ([]acmtypes.Tag, error) {
	return c.certificateClient.ListCertificateTags(ctx, arn)
}

func (c *client) DeleteCertificate(ctx context.Context, arn string) error {
	return c.certificateClient.DeleteCertificate(ctx, arn)
}
