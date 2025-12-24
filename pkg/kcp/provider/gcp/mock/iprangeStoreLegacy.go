package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/servicenetworking/v1"
)

type iprangeStoreLegacy struct {
	mutex           sync.Mutex
	addressStore    *sharedAddressStore    // Use shared address storage
	connectionStore *sharedConnectionStore // Use shared connection storage
}

func (s *iprangeStoreLegacy) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	return s.connectionStore.ListServiceConnections(ctx, projectId, vpcId)
}

func (s *iprangeStoreLegacy) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	return s.connectionStore.CreateServiceConnection(ctx, projectId, vpcId, reservedIpRanges)
}

func (s *iprangeStoreLegacy) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	return s.connectionStore.DeleteServiceConnection(ctx, projectId, vpcId)
}

func (s *iprangeStoreLegacy) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	return s.connectionStore.PatchServiceConnection(ctx, projectId, vpcId, reservedIpRanges)
}

func (s *iprangeStoreLegacy) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	return s.connectionStore.GetServiceNetworkingOperation(ctx, operationName)
}

func (s *iprangeStoreLegacy) ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error) {
	// Use shared store and convert to legacy format
	legacyAddrs, err := s.addressStore.ListGlobalAddressesLegacy(ctx, projectId, vpc)
	if err != nil {
		return nil, err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("ListGlobalAddresses - mock").Info(fmt.Sprintf("Returning %d addresses", len(legacyAddrs)))
	for i, addr := range legacyAddrs {
		if addr != nil {
			logger.WithName("ListGlobalAddresses - mock").Info(fmt.Sprintf("  Address[%d]: Name=%s, Purpose=%s", i, addr.Name, addr.Purpose))
		}
	}

	return &compute.AddressList{Items: legacyAddrs}, nil
}

func (s *iprangeStoreLegacy) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (*compute.Operation, error) {
	// Delegate to shared store
	_, err := s.addressStore.CreatePscIpRange(ctx, projectId, vpcName, name, description, address, prefixLength)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *iprangeStoreLegacy) DeleteIpRange(ctx context.Context, projectId, name string) (*compute.Operation, error) {
	// Delegate to shared store
	_, err := s.addressStore.DeleteIpRange(ctx, projectId, name)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *iprangeStoreLegacy) GetIpRange(ctx context.Context, projectId, name string) (*compute.Address, error) {
	// Use shared store and convert to legacy format
	return s.addressStore.GetIpRangeLegacy(ctx, projectId, name)
}

func (s *iprangeStoreLegacy) GetGlobalOperation(ctx context.Context, projectId, operationName string) (*compute.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return nil, nil
}
