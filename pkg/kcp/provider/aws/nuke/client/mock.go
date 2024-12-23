package client

import (
	"context"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsnfsbackupclient "github.com/kyma-project/cloud-manager/pkg/skr/awsnfsvolumebackup/client"
)

func Mock() awsclient.SkrClientProvider[NukeNfsBackupClient] {
	return func(ctx context.Context, account, region, key, secret, role string) (NukeNfsBackupClient, error) {
		return awsnfsbackupclient.NewMockClient()(ctx, account, region, key, secret, role)
	}
}
