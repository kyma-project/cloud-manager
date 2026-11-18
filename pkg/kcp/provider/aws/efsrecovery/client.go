package efsrecovery

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
)

// EfsClientProvider builds an AWS EFS client for the recovery target account/region. It mirrors
// the credential plumbing used by the NfsInstance reconciler (SkrClientProvider): static default
// creds assume the per-account default role.
type EfsClientProvider func(ctx context.Context, account, region string) (awsclient.EfsClient, error)

func newEfsClientProvider() EfsClientProvider {
	return func(ctx context.Context, account, region string) (awsclient.EfsClient, error) {
		cfg, err := awsclient.NewSkrConfig(
			ctx,
			region,
			awsconfig.AwsConfig.Default.AccessKeyId,
			awsconfig.AwsConfig.Default.SecretAccessKey,
			awsutil.RoleArnDefault(account),
		)
		if err != nil {
			return nil, err
		}
		return awsclient.NewEfsClient(efs.NewFromConfig(cfg)), nil
	}
}
