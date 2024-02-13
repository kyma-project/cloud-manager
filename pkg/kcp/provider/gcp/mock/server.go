package mock

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	nfsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
)

var _ Server = &server{}

func New() Server {
	return &server{
		iprangeStore: &iprangeStore{},
		nfsStore:     &nfsStore{},
	}
}

type server struct {
	*iprangeStore
	*nfsStore
}

func (s *server) ServiceNetworkingClientProvider() client.ClientProvider[iprangeclient.ServiceNetworkingClient] {
	return func(ctx context.Context, saJsonKeyPath string) (iprangeclient.ServiceNetworkingClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP ServiceNetworkingClientProvider mock...")
		return s, nil
	}
}

func (s *server) ComputeClientProvider() client.ClientProvider[iprangeclient.ComputeClient] {
	return func(ctx context.Context, saJsonKeyPath string) (iprangeclient.ComputeClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP ComputeClientProvider mock...")
		return s, nil
	}
}

func (s *server) FilestoreClientProvider() client.ClientProvider[nfsclient.FilestoreClient] {
	return func(ctx context.Context, saJsonKeyPath string) (nfsclient.FilestoreClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP FilestoreClientProvider mock...")
		return s, nil
	}
}
