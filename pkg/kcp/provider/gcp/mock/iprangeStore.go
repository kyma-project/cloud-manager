package mock

import (
	"context"

	"cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/servicenetworking/v1"
	"google.golang.org/protobuf/proto"
)

type iprangeStore struct {
	addressStore    *sharedAddressStore    // Use shared address storage
	connectionStore *sharedConnectionStore // Use shared connection storage
}

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

func (s *iprangeStore) ListGlobalAddresses(ctx context.Context, projectId, vpc string) ([]*computepb.Address, error) {
	return s.addressStore.ListGlobalAddresses(ctx, projectId, vpc)
}

func (s *iprangeStore) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (string, error) {
	return s.addressStore.CreatePscIpRange(ctx, projectId, vpcName, name, description, address, prefixLength)
}

func (s *iprangeStore) DeleteIpRange(ctx context.Context, projectId, name string) (string, error) {
	return s.addressStore.DeleteIpRange(ctx, projectId, name)
}

func (s *iprangeStore) GetIpRange(ctx context.Context, projectId, name string) (*computepb.Address, error) {
	return s.addressStore.GetIpRange(ctx, projectId, name)
}

func (s *iprangeStore) GetGlobalOperation(ctx context.Context, projectId, operationName string) (*computepb.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	// Mock returns a DONE operation
	status := computepb.Operation_DONE
	return &computepb.Operation{
		Name:   proto.String(operationName),
		Status: &status,
	}, nil
}

func (s *iprangeStore) WaitGlobalOperation(ctx context.Context, projectId, operationName string) error {
	if isContextCanceled(ctx) {
		return context.Canceled
	}
	// Mock operations complete immediately
	return nil
}
