package mock

import (
	"context"
	"sync"

	"cloud.google.com/go/redis/apiv1/redispb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/cloudclient"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	backupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	nfsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
	restoreclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	memoryStoreClient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	gcpvpccpeering "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	"google.golang.org/api/googleapi"
)

var _ Server = &server{}

func New() Server {
	return &server{
		iprangeStore:      &iprangeStore{},
		nfsStore:          &nfsStore{},
		serviceUsageStore: &serviceUsageStore{},
		nfsRestoreStore:   &nfsRestoreStore{},
		nfsBackupStore:    &nfsBackupStore{},
		vpcPeeringStore:   &vpcPeeringStore{},
		memoryStoreClientFake: &memoryStoreClientFake{
			mutex:          sync.Mutex{},
			redisInstances: map[string]*redispb.Instance{},
		},
	}
}

type server struct {
	*iprangeStore
	*nfsStore
	*serviceUsageStore
	*nfsRestoreStore
	*nfsBackupStore
	*vpcPeeringStore
	*memoryStoreClientFake
}

func (s *server) SetCreateError(err *googleapi.Error) {
	s.createError = err
}

func (s *server) SetPatchError(err *googleapi.Error) {
	s.patchError = err
}

func (s *server) SetDeleteError(err *googleapi.Error) {
	s.deleteError = err
}

func (s *server) SetGetError(err *googleapi.Error) {
	s.getError = err
}

func (s *server) SetOperationError(err *googleapi.Error) {
	s.operationError = err
}

func (s *server) SetSuEnableError(err *googleapi.Error) {
	s.suEnableError = err
}

func (s *server) SetSuDisableError(err *googleapi.Error) {
	s.suDisableError = err
}

func (s *server) SetSuOperationError(err *googleapi.Error) {
	s.suOperationError = err
}

func (s *server) SetSuIsEnabledError(err *googleapi.Error) {
	s.suIsEnabledError = err
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

func (s *server) FilerestoreClientProvider() client.ClientProvider[restoreclient.FileRestoreClient] {
	return func(ctx context.Context, saJsonKeyPath string) (restoreclient.FileRestoreClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP FilerestoreClientProvider mock...")
		return s, nil
	}
}

func (s *server) FileBackupClientProvider() client.ClientProvider[backupclient.FileBackupClient] {
	return func(ctx context.Context, saJsonKeyPath string) (backupclient.FileBackupClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP FileBackupClientProvider mock...")
		return s, nil
	}
}

func (s *server) VpcPeeringSkrProvider() cloudclient.ClientProvider[gcpvpccpeering.VpcPeeringClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpvpccpeering.VpcPeeringClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP VPCPeeringProvider mock...")
		return s, nil
	}
}

func (s *server) MemoryStoreProviderFake() client.ClientProvider[memoryStoreClient.MemorystoreClient] {
	return func(ctx context.Context, saJsonKeyPath string) (memoryStoreClient.MemorystoreClient, error) {
		return s, nil
	}
}
