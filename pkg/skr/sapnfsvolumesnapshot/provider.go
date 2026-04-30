package sapnfsvolumesnapshot

import (
	"context"

	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
)

func NewSnapshotClientProvider() sapclient.SapClientProvider[sapclient.SnapshotClient] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (sapclient.SnapshotClient, error) {
		f := sapclient.NewClientFactory(pp)
		return f.SnapshotClient(ctx)
	}
}
