package mock

import (
	"context"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	nfsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
	"google.golang.org/api/googleapi"
)

var _ Server = &server{}

func New() Server {
	return &server{
		iprangeStore:      &iprangeStore{},
		nfsStore:          &nfsStore{},
		serviceUsageStore: &serviceUsageStore{},
		nfsRestoreStore:   &nfsRestoreStore{},
	}
}

type server struct {
	*iprangeStore
	*nfsStore
	*serviceUsageStore
	*nfsRestoreStore
}

func (s *server) SetCreateError(error *googleapi.Error) {
	s.createError = error
}

func (s *server) SetPatchError(error *googleapi.Error) {
	s.patchError = error
}

func (s *server) SetDeleteError(error *googleapi.Error) {
	s.deleteError = error
}

func (s *server) SetGetError(error *googleapi.Error) {
	s.getError = error
}

func (s *server) SetOperationError(error *googleapi.Error) {
	s.operationError = error
}

func (s *server) SetSuEnableError(error *googleapi.Error) {
	s.suEnableError = error
}

func (s *server) SetSuDisableError(error *googleapi.Error) {
	s.suDisableError = error
}

func (s *server) SetSuOperationError(error *googleapi.Error) {
	s.suOperationError = error
}

func (s *server) SetSuIsEnabledError(error *googleapi.Error) {
	s.suIsEnabledError = error
}

func (s *server) setRestoreFileError(error *googleapi.Error) {
	s.restoreFileError = error
}

func (s *server) setRestoreOperationError(error *googleapi.Error) {
	s.restoreOperationError = error
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

func (s *server) ServiceUsageClientProvider() client.ClientProvider[client.ServiceUsageClient] {
	return func(ctx context.Context, saJsonKeyPath string) (client.ServiceUsageClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP FilestoreClientProvider mock...")
		return s, nil
	}
}
