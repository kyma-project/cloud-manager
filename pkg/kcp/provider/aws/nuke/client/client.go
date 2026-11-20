package client

import (
	"context"

	acmtypes "github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	backuptypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsnfsvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
	certificateclient "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/certificate/client"
	awscloudresourcesclient "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/cloudresources/client"
	wafpolicyclient "github.com/kyma-project/cloud-manager/pkg/skr/provider/aws/wafpolicy/client"
)

var _ NukeNfsBackupClient = (awsnfsvolumebackupclient.Client)(nil)

// NukeNfsBackupClient is a subset of awsnfsvolumebackupclient.Client
type NukeNfsBackupClient interface {
	CreateBackupVault(ctx context.Context, name string, tags map[string]string) (string, error)
	StartBackupJob(ctx context.Context, params *awsnfsvolumebackupclient.StartBackupJobInput) (*backup.StartBackupJobOutput, error)
	ListBackupVaults(ctx context.Context) ([]backuptypes.BackupVaultListMember, error)
	ListRecoveryPointsForVault(ctx context.Context, accountId, backupVaultName string) ([]backuptypes.RecoveryPointByBackupVault, error)
	DeleteRecoveryPoint(ctx context.Context, backupVaultName, recoveryPointArn string) (*backup.DeleteRecoveryPointOutput, error)
}

var _ NukeWafPolicyClient = (wafpolicyclient.Client)(nil)

// NukeWafPolicyClient is a subset of wafpolicyclient.Client
type NukeWafPolicyClient interface {
	ListWebACLs(ctx context.Context, scope wafv2types.Scope) ([]wafv2types.WebACLSummary, error)
	GetWebACL(ctx context.Context, name, id string, scope wafv2types.Scope) (*wafv2types.WebACL, string, error)
	DeleteWebACL(ctx context.Context, name, id string, scope wafv2types.Scope, lockToken string) error
	ListTagsForWebACL(ctx context.Context, resourceArn string) ([]wafv2types.Tag, error)
	DeleteLoggingConfiguration(ctx context.Context, webAclArn string) error
}

var _ NukeCertificateClient = (certificateclient.Client)(nil)

// NukeCertificateClient is a subset of certificateclient.Client
type NukeCertificateClient interface {
	ListCertificates(ctx context.Context) ([]acmtypes.CertificateSummary, error)
	ListCertificateTags(ctx context.Context, arn string) ([]acmtypes.Tag, error)
	DeleteCertificate(ctx context.Context, arn string) error
}

type NukeWafLogGroup interface {
	awsclient.CloudWatchLogsClient
}

type NukeClient interface {
	NukeNfsBackupClient
	NukeWafPolicyClient
	NukeCertificateClient
	NukeWafLogGroup
}

func NewClientProvider() awsclient.SkrClientProvider[NukeClient] {
	return NukeProvider(
		awsnfsvolumebackupclient.NewClientProvider(),
		wafpolicyclient.NewClientProvider(),
		certificateclient.NewClientProvider(),
		awscloudresourcesclient.NewClientProvider(),
	)
}

func NukeProvider(
	backupProvider awsclient.SkrClientProvider[awsnfsvolumebackupclient.Client],
	webAclProvider awsclient.SkrClientProvider[wafpolicyclient.Client],
	certProvider awsclient.SkrClientProvider[certificateclient.Client],
	wafLogGroupProvider awsclient.SkrClientProvider[awscloudresourcesclient.Client],
) awsclient.SkrClientProvider[NukeClient] {
	return func(ctx context.Context, account, region, key, secret, role string) (NukeClient, error) {
		_backupClient, err := backupProvider(ctx, account, region, key, secret, role)
		if err != nil {
			return nil, err
		}

		_webAclClient, err := webAclProvider(ctx, account, region, key, secret, role)
		if err != nil {
			return nil, err
		}

		_certClient, err := certProvider(ctx, account, region, key, secret, role)
		if err != nil {
			return nil, err
		}

		_wafLogGroupClient, err := wafLogGroupProvider(ctx, account, region, key, secret, role)
		if err != nil {
			return nil, err
		}

		return &client{
			backupClient:      &backupClient{Client: _backupClient},
			webAclClient:      &webAclClient{Client: _webAclClient},
			certClient:        &certClient{Client: _certClient},
			wafLogGroupClient: &wafLogGroupClient{Client: _wafLogGroupClient},
		}, nil
	}
}

type backupClient struct {
	awsnfsvolumebackupclient.Client
}

type webAclClient struct {
	wafpolicyclient.Client
}

type certClient struct {
	certificateclient.Client
}

type wafLogGroupClient struct {
	awscloudresourcesclient.Client
}

type client struct {
	*backupClient
	*webAclClient
	*certClient
	*wafLogGroupClient
}
