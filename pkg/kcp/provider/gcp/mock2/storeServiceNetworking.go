package mock2

import (
	"context"

	"google.golang.org/api/servicenetworking/v1"
)

func (s *store) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	panic("implement me")
}

func (s *store) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	panic("implement me")
}

func (s *store) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	panic("implement me")
}

func (s *store) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	panic("implement me")
}

// Operations ======================================

func (s *store) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	panic("implement me")
}
