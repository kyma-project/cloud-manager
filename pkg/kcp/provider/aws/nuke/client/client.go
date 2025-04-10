package client

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsnfsvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
)

type NukeNfsBackupClient interface {
	awsnfsvolumebackupclient.Client
}

func NewClientProvider() awsclient.SkrClientProvider[NukeNfsBackupClient] {
	return func(ctx context.Context, account, region, key, secret, role string) (NukeNfsBackupClient, error) {
		cfg, err := awsclient.NewSkrConfig(ctx, region, key, secret, role)
		if err != nil {
			return nil, err
		}
		return awsnfsvolumebackupclient.NewClient(backup.NewFromConfig(cfg)), nil
	}
}
