package mock

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/protobuf/proto"
)

// sharedAddressStore provides thread-safe storage for GCP addresses
// Stores addresses in gRPC format (computepb.Address) which is the canonical format
// Provides conversion to Discovery API format (compute.Address) for legacy v2 code
type sharedAddressStore struct {
	mutex     sync.Mutex
	addresses []*computepb.Address
}

// GetIpRange retrieves an address in gRPC format (for new code)
func (s *sharedAddressStore) GetIpRange(ctx context.Context, projectId, name string) (*computepb.Address, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := composed.LoggerFromCtx(ctx)
	id := fmt.Sprintf("projects/%s/address/%s", projectId, name)

	for _, addr := range s.addresses {
		if value, okay := addr.Labels["id"]; okay && value == id {
			logger.WithName("GetIpRange - mock").Info("Got Address. ", "address =", addr)
			return addr, nil
		}
	}

	logger.WithName("GetIpRange - mock").Info(fmt.Sprintf("Length :: %d", len(s.addresses)))
	return nil, &googleapi.Error{
		Code:    404,
		Message: "Not able to find the address",
	}
}

// GetIpRangeLegacy retrieves an address in Discovery API format (for v2 legacy code)
func (s *sharedAddressStore) GetIpRangeLegacy(ctx context.Context, projectId, name string) (*compute.Address, error) {
	grpcAddr, err := s.GetIpRange(ctx, projectId, name)
	if err != nil {
		return nil, err
	}

	// Convert gRPC format to Discovery API format
	return convertGrpcToDiscovery(grpcAddr), nil
}

// ListGlobalAddresses returns all addresses in gRPC format
func (s *sharedAddressStore) ListGlobalAddresses(ctx context.Context, projectId, vpc string) ([]*computepb.Address, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("ListGlobalAddresses - mock").Info(fmt.Sprintf("Length :: %d", len(s.addresses)))

	return s.addresses, nil
}

// ListGlobalAddressesLegacy returns all addresses in Discovery API format
func (s *sharedAddressStore) ListGlobalAddressesLegacy(ctx context.Context, projectId, vpc string) ([]*compute.Address, error) {
	grpcAddrs, err := s.ListGlobalAddresses(ctx, projectId, vpc)
	if err != nil {
		return nil, err
	}

	// Convert gRPC format to Discovery API format
	legacyAddrs := make([]*compute.Address, len(grpcAddrs))
	for i, grpcAddr := range grpcAddrs {
		legacyAddrs[i] = convertGrpcToDiscovery(grpcAddr)
	}
	return legacyAddrs, nil
}

// CreatePscIpRange creates an address (stores in gRPC format internally)
func (s *sharedAddressStore) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := composed.LoggerFromCtx(ctx)

	id := fmt.Sprintf("projects/%s/address/%s", projectId, name)
	lbls := map[string]string{
		"id": id,
	}
	addr := &computepb.Address{
		Name:         proto.String(name),
		Description:  proto.String(description),
		Address:      proto.String(address),
		PrefixLength: proto.Int32(int32(prefixLength)),
		Network:      proto.String(client.GetVPCPath(projectId, vpcName)),
		AddressType:  proto.String(computepb.Address_INTERNAL.String()),
		Purpose:      proto.String(computepb.Address_VPC_PEERING.String()),
		Labels:       lbls,
	}
	s.addresses = append(s.addresses, addr)
	logger.WithName("CreatePscIpRange - mock").Info(fmt.Sprintf("Length :: %d", len(s.addresses)))

	return "", nil
}

// DeleteIpRange removes an address
func (s *sharedAddressStore) DeleteIpRange(ctx context.Context, projectId, name string) (string, error) {
	if isContextCanceled(ctx) {
		return "", context.Canceled
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := composed.LoggerFromCtx(ctx)

	id := fmt.Sprintf("projects/%s/address/%s", projectId, name)
	for i, addr := range s.addresses {
		if value, okay := addr.Labels["id"]; okay && value == id {
			s.addresses = append(s.addresses[:i], s.addresses[i+1:]...)
			break
		}
	}

	logger.WithName("DeleteIpRange - mock").Info(fmt.Sprintf("Length :: %d", len(s.addresses)))
	return "", nil
}

// convertGrpcToDiscovery converts a gRPC Address to Discovery API Address
func convertGrpcToDiscovery(grpc *computepb.Address) *compute.Address {
	if grpc == nil {
		return nil
	}

	return &compute.Address{
		Name:         grpc.GetName(),
		Description:  grpc.GetDescription(),
		Address:      grpc.GetAddress(),
		PrefixLength: int64(grpc.GetPrefixLength()),
		Network:      grpc.GetNetwork(),
		AddressType:  grpc.GetAddressType(),
		Purpose:      grpc.GetPurpose(),
	}
}
