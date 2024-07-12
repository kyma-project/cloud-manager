package mock

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/servicenetworking/v1"
)

type iprangeStore struct {
	connections []*servicenetworking.Connection
	addresses   []*compute.Address
}

func (s *iprangeStore) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("ListServiceConnections - mock").Info(fmt.Sprintf("Length :: %d", len(s.connections)))
	return s.connections, nil
}

func (s *iprangeStore) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	conn := servicenetworking.Connection{
		Network:               client.GetVPCPath(projectId, vpcId),
		ReservedPeeringRanges: reservedIpRanges,
		Peering:               client.PsaPeeringName,
		Service:               client.ServiceNetworkingServicePath,
	}
	s.connections = append(s.connections, &conn)
	logger.WithName("CreateServiceConnection - mock").Info("SrvcConnection", "List", s.connections)

	return nil, nil
}

func (s *iprangeStore) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)
	nw := client.GetVPCPath(projectId, vpcId)
	for i, conn := range s.connections {
		if conn != nil && conn.Network == nw {
			s.connections = append(s.connections[:i], s.connections[i+1:]...)
			break
		}
	}
	logger.WithName("DeleteServiceConnection - mock").Info(fmt.Sprintf("Length :: %d", len(s.connections)))

	return nil, nil
}

func (s *iprangeStore) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("PatchServiceConnection - mock").Info(fmt.Sprintf("Length :: %d", len(s.connections)))

	nw := client.GetVPCPath(projectId, vpcId)
	for i, conn := range s.connections {
		if conn != nil && conn.Network == nw {
			s.connections[i].ReservedPeeringRanges = reservedIpRanges
			break
		}
	}
	return nil, nil
}

func (s *iprangeStore) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return nil, nil
}

func (s *iprangeStore) ListGlobalAddresses(ctx context.Context, projectId, vpc string) (*compute.AddressList, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)
	logger.WithName("ListGlobalAddresses - mock").Info(fmt.Sprintf("Length :: %d", len(s.addresses)))

	return &compute.AddressList{Items: s.addresses}, nil
}

func (s *iprangeStore) CreatePscIpRange(ctx context.Context, projectId, vpcName, name, description, address string, prefixLength int64) (*compute.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	id := fmt.Sprintf("projects/%s/address/%s", projectId, name)
	lbls := map[string]string{
		"id": id,
	}
	addr := compute.Address{
		Name:         name,
		Description:  description,
		Address:      address,
		PrefixLength: prefixLength,
		Network:      client.GetVPCPath(projectId, vpcName),
		AddressType:  string(client.AddressTypeInternal),
		Purpose:      string(client.IpRangePurposeVPCPeering),
		Labels:       lbls,
	}
	s.addresses = append(s.addresses, &addr)
	logger.WithName("CreatePscIpRange - mock").Info(fmt.Sprintf("Length :: %d", len(s.addresses)))

	return nil, nil
}

func (s *iprangeStore) DeleteIpRange(ctx context.Context, projectId, name string) (*compute.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	id := fmt.Sprintf("projects/%s/address/%s", projectId, name)
	for i, addr := range s.addresses {
		if value, okay := addr.Labels["id"]; okay && value == id {
			s.addresses = append(s.addresses[:i], s.addresses[i+1:]...)
			break
		}
	}

	logger.WithName("DeleteIpRange - mock").Info(fmt.Sprintf("Length :: %d", len(s.addresses)))
	return nil, nil
}

func (s *iprangeStore) GetIpRange(ctx context.Context, projectId, name string) (*compute.Address, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	logger := composed.LoggerFromCtx(ctx)

	var result *compute.Address
	id := fmt.Sprintf("projects/%s/address/%s", projectId, name)

	for _, addr := range s.addresses {
		if value, okay := addr.Labels["id"]; okay && value == id {
			result = addr
			logger.WithName("GetIpRange - mock").Info("Got Address. ", "address =", addr)
			return result, nil
		}
	}

	logger.WithName("GetIpRange - mock").Info(fmt.Sprintf("Length :: %d", len(s.addresses)))
	return nil, &googleapi.Error{
		Code:    404,
		Message: "Not able to find the address",
	}
}

func (s *iprangeStore) GetGlobalOperation(ctx context.Context, projectId, operationName string) (*compute.Operation, error) {
	if isContextCanceled(ctx) {
		return nil, context.Canceled
	}

	return nil, nil
}
