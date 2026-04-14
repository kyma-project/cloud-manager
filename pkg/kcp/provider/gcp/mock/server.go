package mock

import (
	"context"
	"sync"

	iprangeallocate "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/allocate"
	gcpexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/exposedData/client"
	"github.com/kyma-project/cloud-manager/pkg/util"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/iprange/client"
	gcpnfsbackupclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsbackup/client/v1"
	gcpnfsinstancev1client "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsinstance/v1/client"
	gcpnfsrestoreclientv1 "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/nfsrestore/client/v1"
	gcpvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcpeering/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/servicenetworking/v1"
)

var _ Server = &server{}

func New() Server {

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
		nfsStore:          &nfsStore{},
		serviceUsageStore: &serviceUsageStore{},
		nfsRestoreStore:   &nfsRestoreStore{},
		nfsBackupStore:    &nfsBackupStore{},
		vpcPeeringStore:   &vpcPeeringStore{},
		exposedDataStore: &exposedDataStore{
			ipPool: util.Must(iprangeallocate.NewAddressSpace("33.0.0.0/16")),
		},
	}
}

type server struct {
	iprangeStore       *iprangeStore
	iprangeStoreLegacy *iprangeStoreLegacy
	*nfsStore
	*serviceUsageStore
	*nfsRestoreStore
	*nfsBackupStore
	*vpcPeeringStore
	*exposedDataStore
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
		// Return the legacy store directly - it already implements OldComputeClient interface with Discovery API types
		return s.iprangeStoreLegacy, nil
	}
}

func (s *server) FilestoreClientProvider() client.ClientProvider[gcpnfsinstancev1client.FilestoreClient] {
	return func(ctx context.Context, credentialsFile string) (gcpnfsinstancev1client.FilestoreClient, error) {
		return s, nil
	}
}

func (s *server) ServiceUsageClientProvider() client.ClientProvider[client.ServiceUsageClient] {
	return func(ctx context.Context, credentialsFile string) (client.ServiceUsageClient, error) {
		return s, nil
	}
}

func (s *server) FilerestoreClientProvider() client.ClientProvider[gcpnfsrestoreclientv1.FileRestoreClient] {
	return func(ctx context.Context, credentialsFile string) (gcpnfsrestoreclientv1.FileRestoreClient, error) {
		return s.nfsRestoreStore, nil
	}
}

func (s *server) FileBackupClientProvider() client.ClientProvider[gcpnfsbackupclientv1.FileBackupClient] {
	return func(ctx context.Context, credentialsFile string) (gcpnfsbackupclientv1.FileBackupClient, error) {
		return s, nil
	}
}

func (s *server) VpcPeeringProvider() client.GcpClientProvider[gcpvpcpeeringclient.VpcPeeringClient] {
	return func(_ string) gcpvpcpeeringclient.VpcPeeringClient {
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

func (s *server) GetGlobalAddress(ctx context.Context, req *computepb.GetGlobalAddressRequest, opts ...gax.CallOption) (*computepb.Address, error) {
	return s.iprangeStore.GetGlobalAddress(ctx, req, opts...)
}

func (s *server) DeleteGlobalAddress(ctx context.Context, req *computepb.DeleteGlobalAddressRequest, opts ...gax.CallOption) (client.VoidOperation, error) {
	return s.iprangeStore.DeleteGlobalAddress(ctx, req, opts...)
}

func (s *server) InsertGlobalAddress(ctx context.Context, req *computepb.InsertGlobalAddressRequest, opts ...gax.CallOption) (client.VoidOperation, error) {
	return s.iprangeStore.InsertGlobalAddress(ctx, req, opts...)
}

func (s *server) ListGlobalAddresses(ctx context.Context, req *computepb.ListGlobalAddressesRequest, opts ...gax.CallOption) client.Iterator[*computepb.Address] {
	return s.iprangeStore.ListGlobalAddresses(ctx, req, opts...)
}

func (s *server) GetComputeGlobalOperation(ctx context.Context, req *computepb.GetGlobalOperationRequest, opts ...gax.CallOption) (*computepb.Operation, error) {
	return s.iprangeStore.GetComputeGlobalOperation(ctx, req, opts...)
}

func (s *server) ListComputeGlobalOperations(ctx context.Context, req *computepb.ListGlobalOperationsRequest, opts ...gax.CallOption) client.Iterator[*computepb.Operation] {
	return s.iprangeStore.ListComputeGlobalOperations(ctx, req, opts...)
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
