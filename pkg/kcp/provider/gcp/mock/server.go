package mock

import (
	"context"
	"sync"

	iprangeallocate "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	"github.com/kyma-project/cloud-manager/pkg/util"

	"cloud.google.com/go/compute/apiv1/computepb"
	"cloud.google.com/go/networkconnectivity/apiv1/networkconnectivitypb"
	"cloud.google.com/go/redis/apiv1/redispb"
	"cloud.google.com/go/redis/cluster/apiv1/clusterpb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpnfsbackupclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v1"
	gcpnfsbackupclientv2 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v2"
	gcpnfsinstancev1client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v1/client"
	gcpnfsinstancev2client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v2/client"
	gcpnfsrestoreclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client"
	gcpredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/rediscluster/client"
	gcpredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/redisinstance/client"
	gcpsubnetclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/subnet/client"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/servicenetworking/v1"
)

var _ Server = &server{}

func New() Server {

	regionalOperationsClientfake := &regionalOperationsClientFake{
		mutex:      sync.Mutex{},
		operations: map[string]*computepb.Operation{},
	}

	// Create shared address storage that both iprangeStore and iprangeStoreLegacy can use
	// Thread-safety: sharedAddressStore has its own mutex and protects all access to addresses.
	// Both iprangeStore and iprangeStoreLegacy share the same instance, ensuring consistent state
	// across legacy (v2) and refactored code paths. Concurrent access is safe because the shared
	// store's mutex serializes all operations.
	sharedAddresses := &sharedAddressStore{
		mutex:     sync.Mutex{},
		addresses: []*computepb.Address{},
	}

	// Create shared connection storage that both iprangeStore and iprangeStoreLegacy can use
	// Thread-safety: sharedConnectionStore has its own mutex and protects all access to connections.
	// Both iprangeStore and iprangeStoreLegacy share the same instance for consistent PSA connection state.
	// Concurrent access is safe because the shared store's mutex serializes all operations.
	sharedConnections := &sharedConnectionStore{
		mutex:       sync.Mutex{},
		connections: []*servicenetworking.Connection{},
	}

	return &server{
		iprangeStore: &iprangeStore{
			addressStore:    sharedAddresses,
			connectionStore: sharedConnections,
		},
		iprangeStoreLegacy: &iprangeStoreLegacy{
			addressStore:    sharedAddresses,
			connectionStore: sharedConnections,
		},
		computeClientFake: &computeClientFake{
			mutex:                 sync.Mutex{},
			subnets:               map[string]*computepb.Subnetwork{},
			operationsClientUtils: regionalOperationsClientfake,
		},
		networkConnectivityClientFake: &networkConnectivityClientFake{
			mutex:              sync.Mutex{},
			connectionPolicies: map[string]*networkconnectivitypb.ServiceConnectionPolicy{},
		},
		nfsStore:              &nfsStore{},
		filestoreClientFakeV2: newFilestoreClientFakeV2(),
		serviceUsageStore:     &serviceUsageStore{},
		nfsRestoreStore:       &nfsRestoreStore{},
		nfsBackupStore:        &nfsBackupStore{},
		nfsBackupStoreV2:      &nfsBackupStoreV2{},
		vpcPeeringStore:       &vpcPeeringStore{},
		memoryStoreClientFake: &memoryStoreClientFake{
			mutex:          sync.Mutex{},
			redisInstances: map[string]*redispb.Instance{},
		},
		memoryStoreClusterClientFake: &memoryStoreClusterClientFake{
			mutex:         sync.Mutex{},
			redisClusters: map[string]*clusterpb.Cluster{},
		},
		exposedDataStore: &exposedDataStore{
			ipPool: util.Must(iprangeallocate.NewAddressSpace("33.0.0.0/16")),
		},
		regionalOperationsClientFake: regionalOperationsClientfake,
	}
}

type server struct {
	iprangeStore       *iprangeStore
	iprangeStoreLegacy *iprangeStoreLegacy
	*computeClientFake
	*networkConnectivityClientFake
	*nfsStore
	*filestoreClientFakeV2
	*serviceUsageStore
	*nfsRestoreStore
	*nfsBackupStore
	*nfsBackupStoreV2
	*vpcPeeringStore
	*memoryStoreClientFake
	*memoryStoreClusterClientFake
	*exposedDataStore
	*regionalOperationsClientFake
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
	return func(ctx context.Context, credentialsFile string) (gcpiprangeclient.ServiceNetworkingClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP ServiceNetworkingClientProvider mock...")
		// Return the legacy store for v2 - it implements ServiceNetworkingClient with Discovery API types
		return s.iprangeStoreLegacy, nil
	}
}

// ServiceNetworkingClientProviderGcp returns a GcpClientProvider (NEW pattern)
func (s *server) ServiceNetworkingClientProviderGcp() client.GcpClientProvider[gcpiprangeclient.ServiceNetworkingClient] {
	return func(_ string) gcpiprangeclient.ServiceNetworkingClient {
		// For NEW pattern (refactored), return the new gRPC-based store
		return s.iprangeStore
	}
}

func (s *server) ComputeClientProvider() client.ClientProvider[gcpiprangeclient.ComputeClient] {
	return func(ctx context.Context, credentialsFile string) (gcpiprangeclient.ComputeClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP ComputeClientProvider mock...")
		// For NEW pattern (refactored), return the new gRPC-based store
		return s.iprangeStore, nil
	}
}

// ComputeClientProviderGcp returns a GcpClientProvider (NEW pattern)
func (s *server) ComputeClientProviderGcp() client.GcpClientProvider[gcpiprangeclient.ComputeClient] {
	return func(_ string) gcpiprangeclient.ComputeClient {
		// For NEW pattern (refactored), return the new gRPC-based store
		return s.iprangeStore
	}
}

// OldComputeClientProvider returns OLD-style ClientProvider for v2 legacy tests
// Returns an adapter that converts between gRPC types (mock) and Discovery API types (v2)
func (s *server) OldComputeClientProvider() client.ClientProvider[gcpiprangeclient.OldComputeClient] {
	return func(ctx context.Context, credentialsFile string) (gcpiprangeclient.OldComputeClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP OldComputeClientProvider mock...")
		// Return the legacy store directly - it already implements OldComputeClient interface with Discovery API types
		return s.iprangeStoreLegacy, nil
	}
}

func (s *server) SubnetComputeClientProvider() client.GcpClientProvider[gcpsubnetclient.ComputeClient] {
	return func(_ string) gcpsubnetclient.ComputeClient {
		return s
	}
}

func (s *server) SubnetRegionOperationsClientProvider() client.GcpClientProvider[gcpsubnetclient.RegionOperationsClient] {
	return func(_ string) gcpsubnetclient.RegionOperationsClient {
		return s
	}
}

func (s *server) SubnetNetworkConnectivityProvider() client.GcpClientProvider[gcpsubnetclient.NetworkConnectivityClient] {
	return func(_ string) gcpsubnetclient.NetworkConnectivityClient {
		return s
	}
}

func (s *server) FilestoreClientProvider() client.ClientProvider[gcpnfsinstancev1client.FilestoreClient] {
	return func(ctx context.Context, credentialsFile string) (gcpnfsinstancev1client.FilestoreClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP FilestoreClientProvider mock...")
		return s, nil
	}
}

func (s *server) FilestoreClientProviderV2() client.GcpClientProvider[gcpnfsinstancev2client.FilestoreClient] {
	return func(_ string) gcpnfsinstancev2client.FilestoreClient {
		return s
	}
}

func (s *server) ServiceUsageClientProvider() client.ClientProvider[client.ServiceUsageClient] {
	return func(ctx context.Context, credentialsFile string) (client.ServiceUsageClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP FilestoreClientProvider mock...")
		return s, nil
	}
}

func (s *server) FilerestoreClientProvider() client.ClientProvider[gcpnfsrestoreclient.FileRestoreClient] {
	return func(ctx context.Context, credentialsFile string) (gcpnfsrestoreclient.FileRestoreClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP FilerestoreClientProvider mock...")
		return s, nil
	}
}

func (s *server) FileBackupClientProvider() client.ClientProvider[gcpnfsbackupclientv1.FileBackupClient] {
	return func(ctx context.Context, credentialsFile string) (gcpnfsbackupclientv1.FileBackupClient, error) {
		logger := composed.LoggerFromCtx(ctx)
		logger.Info("Inside the GCP FileBackupClientProvider mock...")
		return s, nil
	}
}

func (s *server) FileBackupClientProviderV2() client.GcpClientProvider[gcpnfsbackupclientv2.FileBackupClient] {
	return func() gcpnfsbackupclientv2.FileBackupClient {
		return s.nfsBackupStoreV2
	}
}

func (s *server) VpcPeeringProvider() client.GcpClientProvider[gcpvpcpeeringclient.VpcPeeringClient] {
	return func(_ string) gcpvpcpeeringclient.VpcPeeringClient {
		return s
	}
}

func (s *server) MemoryStoreProviderFake() client.GcpClientProvider[gcpredisinstanceclient.MemorystoreClient] {
	return func(_ string) gcpredisinstanceclient.MemorystoreClient {
		return s
	}
}

func (s *server) MemoryStoreClusterProviderFake() client.GcpClientProvider[gcpredisclusterclient.MemorystoreClusterClient] {
	return func(_ string) gcpredisclusterclient.MemorystoreClusterClient {
		return s
	}
}

func (s *server) ExposedDataProvider() client.GcpClientProvider[gcpexposeddataclient.Client] {
	return func(_ string) gcpexposeddataclient.Client {
		return s
	}
}

// Delegation methods for IpRangeClient interface (NEW pattern - refactored)
// These delegate to iprangeStore which uses gRPC types

func (s *server) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (string, error) {
	return s.iprangeStore.CreatePscIpRange(ctx, projectId, vpcName, name, description, address, prefixLength)
}

func (s *server) DeleteIpRange(ctx context.Context, projectId, name string) (string, error) {
	return s.iprangeStore.DeleteIpRange(ctx, projectId, name)
}

func (s *server) GetIpRange(ctx context.Context, projectId, name string) (*computepb.Address, error) {
	return s.iprangeStore.GetIpRange(ctx, projectId, name)
}

func (s *server) ListGlobalAddresses(ctx context.Context, projectId, vpc string) ([]*computepb.Address, error) {
	return s.iprangeStore.ListGlobalAddresses(ctx, projectId, vpc)
}

func (s *server) GetGlobalOperation(ctx context.Context, projectId, operationName string) (*computepb.Operation, error) {
	return s.iprangeStore.GetGlobalOperation(ctx, projectId, operationName)
}

func (s *server) WaitGlobalOperation(ctx context.Context, projectId, operationName string) error {
	return s.iprangeStore.WaitGlobalOperation(ctx, projectId, operationName)
}

func (s *server) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	return s.iprangeStore.CreateServiceConnection(ctx, projectId, vpcId, reservedIpRanges)
}

func (s *server) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	return s.iprangeStore.DeleteServiceConnection(ctx, projectId, vpcId)
}

func (s *server) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	return s.iprangeStore.PatchServiceConnection(ctx, projectId, vpcId, reservedIpRanges)
}

func (s *server) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	return s.iprangeStore.ListServiceConnections(ctx, projectId, vpcId)
}

func (s *server) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	return s.iprangeStore.GetServiceNetworkingOperation(ctx, operationName)
}

// IpRangeClientUtils implementation - for test assertions with Discovery API format

func (s *server) GetIpRangeDiscovery(ctx context.Context, projectId, name string) (*compute.Address, error) {
	return s.iprangeStoreLegacy.GetIpRange(ctx, projectId, name)
}

func (s *server) ListGlobalAddressesDiscovery(ctx context.Context, projectId, vpc string) ([]*compute.Address, error) {
	list, err := s.iprangeStoreLegacy.ListGlobalAddresses(ctx, projectId, vpc)
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}
