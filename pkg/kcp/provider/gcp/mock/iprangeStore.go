package mock

import (
	"context"
	"iter"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/googleapis/gax-go/v2"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/servicenetworking/v1"
	"google.golang.org/protobuf/proto"
)

type iprangeStore struct {
	addressStore    *sharedAddressStore    // Use shared address storage
	connectionStore *sharedConnectionStore // Use shared connection storage
}

// ServiceNetworkingClient methods

func (s *iprangeStore) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	return s.connectionStore.ListServiceConnections(ctx, projectId, vpcId)
}

func (s *iprangeStore) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	return s.connectionStore.CreateServiceConnection(ctx, projectId, vpcId, reservedIpRanges)
}

func (s *iprangeStore) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	return s.connectionStore.DeleteServiceConnection(ctx, projectId, vpcId)
}

func (s *iprangeStore) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	return s.connectionStore.PatchServiceConnection(ctx, projectId, vpcId, reservedIpRanges)
}

func (s *iprangeStore) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	return s.connectionStore.GetServiceNetworkingOperation(ctx, operationName)
}

// gcpclient.GlobalAddressesClient methods

func (s *iprangeStore) GetGlobalAddress(ctx context.Context, req *computepb.GetGlobalAddressRequest, _ ...gax.CallOption) (*computepb.Address, error) {
	return s.addressStore.GetIpRange(ctx, req.Project, req.Address)
}

func (s *iprangeStore) DeleteGlobalAddress(ctx context.Context, req *computepb.DeleteGlobalAddressRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	opName, err := s.addressStore.DeleteIpRange(ctx, req.Project, req.Address)
	if err != nil {
		return nil, err
	}
	return &mockVoidOperation{name: opName}, nil
}

func (s *iprangeStore) InsertGlobalAddress(ctx context.Context, req *computepb.InsertGlobalAddressRequest, _ ...gax.CallOption) (gcpclient.VoidOperation, error) {
	addr := req.AddressResource
	opName, err := s.addressStore.CreatePscIpRange(ctx, req.Project,
		"", // vpcName not needed since address resource has full network path
		addr.GetName(), addr.GetDescription(), addr.GetAddress(), int64(addr.GetPrefixLength()))
	if err != nil {
		return nil, err
	}
	return &mockVoidOperation{name: opName}, nil
}

func (s *iprangeStore) ListGlobalAddresses(ctx context.Context, req *computepb.ListGlobalAddressesRequest, _ ...gax.CallOption) gcpclient.Iterator[*computepb.Address] {
	addrs, err := s.addressStore.ListGlobalAddresses(ctx, req.Project, "")
	return &mockAddressIterator{addresses: addrs, err: err}
}

// gcpclient.ComputeGlobalOperationsClient methods

func (s *iprangeStore) GetComputeGlobalOperation(ctx context.Context, req *computepb.GetGlobalOperationRequest, _ ...gax.CallOption) (*computepb.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	// Mock returns a DONE operation
	status := computepb.Operation_DONE
	return &computepb.Operation{
		Name:   proto.String(req.Operation),
		Status: &status,
	}, nil
}

func (s *iprangeStore) ListComputeGlobalOperations(ctx context.Context, req *computepb.ListGlobalOperationsRequest, _ ...gax.CallOption) gcpclient.Iterator[*computepb.Operation] {
	return &mockOperationIterator{}
}

// CreatePscIpRange is the value-add method kept on the ComputeClient interface
func (s *iprangeStore) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (string, error) {
	return s.addressStore.CreatePscIpRange(ctx, projectId, vpcName, name, description, address, prefixLength)
}

// mockVoidOperation implements gcpclient.VoidOperation for the old mock
type mockVoidOperation struct {
	name string
}

func (o *mockVoidOperation) Name() string                                      { return o.name }
func (o *mockVoidOperation) Done() bool                                        { return true }
func (o *mockVoidOperation) Poll(_ context.Context, _ ...gax.CallOption) error { return nil }
func (o *mockVoidOperation) Wait(_ context.Context, _ ...gax.CallOption) error { return nil }

// mockAddressIterator implements gcpclient.Iterator[*computepb.Address] for the old mock
type mockAddressIterator struct {
	addresses []*computepb.Address
	pos       int
	err       error
}

func (it *mockAddressIterator) Next() (*computepb.Address, error) {
	if it.err != nil {
		return nil, it.err
	}
	if it.pos >= len(it.addresses) {
		return nil, nil
	}
	addr := it.addresses[it.pos]
	it.pos++
	return addr, nil
}

func (it *mockAddressIterator) All() iter.Seq2[*computepb.Address, error] {
	return func(yield func(*computepb.Address, error) bool) {
		if it.err != nil {
			yield(nil, it.err)
			return
		}
		for _, addr := range it.addresses {
			if !yield(addr, nil) {
				return
			}
		}
	}
}

// mockOperationIterator implements gcpclient.Iterator[*computepb.Operation] for the old mock
type mockOperationIterator struct{}

func (it *mockOperationIterator) Next() (*computepb.Operation, error) {
	return nil, nil
}

func (it *mockOperationIterator) All() iter.Seq2[*computepb.Operation, error] {
	return func(yield func(*computepb.Operation, error) bool) {}
}
