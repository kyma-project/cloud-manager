package mock

import (
	"context"
	"fmt"
	"sync"

	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/servicenetworking/v1"
)

// sharedConnectionStore provides centralized storage for PSA connections
// Used by both iprangeStore (NEW/refactored) and iprangeStoreLegacy (v2)
type sharedConnectionStore struct {
	mutex       sync.Mutex
	connections []*servicenetworking.Connection
}

func (s *sharedConnectionStore) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := composed.LoggerFromCtx(ctx)

	// Filter connections by network path (project + VPC)
	networkPath := client.GetVPCPath(projectId, vpcId)
	var matchingConnections []*servicenetworking.Connection
	for _, conn := range s.connections {
		if conn != nil && conn.Network == networkPath {
			matchingConnections = append(matchingConnections, conn)
		}
	}

	logger.WithName("ListServiceConnections - mock").Info(fmt.Sprintf("Length :: %d", len(s.connections)))
	logger.WithName("ListServiceConnections - mock").Info(fmt.Sprintf("Returning %d connections for network %s", len(matchingConnections), networkPath))
	for i, conn := range matchingConnections {
		logger.WithName("ListServiceConnections - mock").Info(fmt.Sprintf("  Connection[%d]: ReservedPeeringRanges=%v", i, conn.ReservedPeeringRanges))
	}
	return matchingConnections, nil
}

func (s *sharedConnectionStore) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := composed.LoggerFromCtx(ctx)

	// Make a copy of the slice to avoid reference issues
	ipRangesCopy := make([]string, len(reservedIpRanges))
	copy(ipRangesCopy, reservedIpRanges)

	conn := servicenetworking.Connection{
		Network:               client.GetVPCPath(projectId, vpcId),
		Peering:               client.PsaPeeringName,
		ReservedPeeringRanges: ipRangesCopy,
		Service:               client.ServiceNetworkingServicePath,
	}
	s.connections = append(s.connections, &conn)

	logger.WithName("CreateServiceConnection - mock").Info("SrvcConnection", "List", s.connections)

	// Return nil for synchronous completion (v2 code checks: if operation != nil, set OpIdentifier and requeue)
	return nil, nil
}

func (s *sharedConnectionStore) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("PatchServiceConnection - mock").Info(fmt.Sprintf("Length :: %d", len(s.connections)))

	// Make a copy of the slice to avoid reference issues
	ipRangesCopy := make([]string, len(reservedIpRanges))
	copy(ipRangesCopy, reservedIpRanges)

	nw := client.GetVPCPath(projectId, vpcId)
	var matchingConnection *servicenetworking.Connection
	for _, conn := range s.connections {
		if conn != nil && conn.Network == nw {
			matchingConnection = conn
			break
		}
	}

	if matchingConnection == nil {
		logger.WithName("PatchServiceConnection - mock").Info("No matching connection found to update!")
		return nil, nil
	}

	matchingConnection.ReservedPeeringRanges = ipRangesCopy
	logger.WithName("PatchServiceConnection - mock").Info(fmt.Sprintf("Updated connection with ReservedPeeringRanges: %v", ipRangesCopy))

	// Return nil for synchronous completion
	return nil, nil
}

func (s *sharedConnectionStore) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	logger := composed.LoggerFromCtx(ctx)

	nw := client.GetVPCPath(projectId, vpcId)
	for i, conn := range s.connections {
		if conn != nil && conn.Network == nw {
			s.connections = append(s.connections[:i], s.connections[i+1:]...)
			break
		}
	}

	logger.WithName("DeleteServiceConnection - mock").Info(fmt.Sprintf("Length :: %d", len(s.connections)))

	// Return nil for synchronous completion
	return nil, nil
}

func (s *sharedConnectionStore) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return nil, nil
}
