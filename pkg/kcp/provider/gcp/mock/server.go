package mock

import (
	"context"
	"sync"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"cloud.google.com/go/redis/apiv1/redispb"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpnfsbackupclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client"
	gcpnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/client"
	gcpnfsrestoreclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	gcpredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
	gcpredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	gcpsubnetclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	"google.golang.org/api/googleapi"
)

var _ Server = &server{}

func New() Server {
	return &server{
		iprangeStore: &iprangeStore{},
		computeClientFake: &computeClientFake{
			mutex:   sync.Mutex{},
			subnets: map[string]*computepb.Subnetwork{},
		},
		networkConnectivityClientFake: &networkConnectivityClientFake{
			mutex:              sync.Mutex{},
			connectionPolicies: map[string]*networkconnectivitypb.ServiceConnectionPolicy{},
		},
		nfsStore:          &nfsStore{},
		serviceUsageStore: &serviceUsageStore{},
		nfsRestoreStore:   &nfsRestoreStore{},
		nfsBackupStore:    &nfsBackupStore{},
		vpcPeeringStore:   &vpcPeeringStore{},
		memoryStoreClientFake: &memoryStoreClientFake{
			mutex:          sync.Mutex{},
			redisInstances: map[string]*redispb.Instance{},
		},
		memoryStoreClusterClientFake: &memoryStoreClusterClientFake{
			mutex:         sync.Mutex{},
			redisClusters: map[string]*clusterpb.Cluster{},
		},
	}
}

type server struct {
	*iprangeStore
	*computeClientFake
	*networkConnectivityClientFake
	*nfsStore
	*serviceUsageStore
	*nfsRestoreStore
	*nfsBackupStore
	*vpcPeeringStore
	*memoryStoreClientFake
	*memoryStoreClusterClientFake
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

func (s *server) ServiceNetworkingClientProvider() client.ClientProvider[gcpiprangeclient.ServiceNetworkingClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpiprangeclient.ServiceNetworkingClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP ServiceNetworkingClientProvider mock...")
		return s, nil
	}
}

func (s *server) ComputeClientProvider() client.ClientProvider[gcpiprangeclient.ComputeClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpiprangeclient.ComputeClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP ComputeClientProvider mock...")
		return s, nil
	}
}

func (s *server) SubnetComputeClientProvider() client.ClientProvider[gcpsubnetclient.ComputeClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpsubnetclient.ComputeClient, error) {
		return s, nil
	}
}
func (s *server) SubnetNetworkConnectivityProvider() client.ClientProvider[gcpsubnetclient.NetworkConnectivityClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpsubnetclient.NetworkConnectivityClient, error) {
		return s, nil
	}
}

func (s *server) FilestoreClientProvider() client.ClientProvider[gcpnfsinstanceclient.FilestoreClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpnfsinstanceclient.FilestoreClient, error) {
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

func (s *server) FilerestoreClientProvider() client.ClientProvider[gcpnfsrestoreclient.FileRestoreClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpnfsrestoreclient.FileRestoreClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP FilerestoreClientProvider mock...")
		return s, nil
	}
}

func (s *server) FileBackupClientProvider() client.ClientProvider[gcpnfsbackupclient.FileBackupClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpnfsbackupclient.FileBackupClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP FileBackupClientProvider mock...")
		return s, nil
	}
}

func (s *server) VpcPeeringProvider() client.ClientProvider[gcpvpcpeeringclient.VpcPeeringClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpvpcpeeringclient.VpcPeeringClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP VPCPeeringProvider mock...")
		return s, nil
	}
}

func (s *server) MemoryStoreProviderFake() client.ClientProvider[gcpredisinstanceclient.MemorystoreClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpredisinstanceclient.MemorystoreClient, error) {
		return s, nil
	}
}

func (s *server) MemoryStoreClusterProviderFake() client.ClientProvider[gcpredisclusterclient.MemorystoreClusterClient] {
	return func(ctx context.Context, saJsonKeyPath string) (gcpredisclusterclient.MemorystoreClusterClient, error) {
		return s, nil
	}
}
