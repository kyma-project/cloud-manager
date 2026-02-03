package client

import (
	"context"

	sapclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/sap/client"
)

type Client interface {
	sapclient.NetworkClient
	sapclient.SubnetClient
	sapclient.RouterClient
}

func NewClientProvider() sapclient.SapClientProvider[Client] {
	return func(ctx context.Context, pp sapclient.ProviderParams) (Client, error) {
		f := sapclient.NewClientFactory(pp)
		nc, err := f.NetworkClient(ctx)
		if err != nil {
			return nil, err
		}
		sc, err := f.SubnetClient(ctx)
		if err != nil {
			return nil, err
		}
		rc, err := f.RouterClient(ctx)
		if err != nil {
			return nil, err
		}
		return &client{
			NetworkClient: nc,
			SubnetClient:  sc,
			RouterClient:  rc,
		}, nil
	}
}

var _ Client = (*client)(nil)

type client struct {
	sapclient.NetworkClient
	sapclient.SubnetClient
	sapclient.RouterClient
}
