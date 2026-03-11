package mock2

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/client"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcputil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/util"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"google.golang.org/api/servicenetworking/v1"
)

func (s *store) PatchServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	netNd := gcputil.NewGlobalNetworkName(projectId, vpcId)

	_, err := s.getNetworkNoLock(projectId, vpcId)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("network %s not found", netNd.String())
	}

	var reservedPeeringRanges []string
	for _, reservedIpRange := range reservedIpRanges {
		// reservedIpRange is the name of the address
		addrNd := gcputil.NewGlobalAddressName(projectId, reservedIpRange)
		addr, found := s.addresses.FindByName(addrNd)
		if !found {
			return nil, gcpmeta.NewBadRequestError("address %s not found", addrNd)
		}
		reservedPeeringRanges = append(reservedPeeringRanges, addr.GetSelfLink())
	}

	con, found := s.serviceConnections.FindByName(netNd)
	if !found {
		return nil, gcpmeta.NewNotFoundError("service connection in network %s not found", netNd)
	}

	con.ReservedPeeringRanges = reservedPeeringRanges

	// don't know the exact operation name format, this will do
	opName := s.newLongRunningOperationName(projectId)
	op := &servicenetworking.Operation{
		Done: true,
		Name: opName.String(),
	}
	s.serviceNetworkingOperations.Add(op, opName)
	return op, nil
}

func (s *store) DeleteServiceConnection(ctx context.Context, projectId, vpcId string) (*servicenetworking.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	netNd := gcputil.NewGlobalNetworkName(projectId, vpcId)

	s.serviceConnections = s.serviceConnections.FilterNotByCallback(func(item FilterableListItem[*servicenetworking.Connection]) bool {
		return netNd.EqualString(item.Obj.Network)
	})

	// don't know the exact operation name format, this will do
	opName := s.newLongRunningOperationName(projectId)
	op := &servicenetworking.Operation{
		Done: true,
		Name: opName.String(),
	}
	s.serviceNetworkingOperations.Add(op, opName)
	return op, nil
}

func (s *store) ListServiceConnections(ctx context.Context, projectId, vpcId string) ([]*servicenetworking.Connection, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	netNd := gcputil.NewGlobalNetworkName(projectId, vpcId)

	var result []*servicenetworking.Connection
	for _, item := range s.serviceConnections.items {
		if netNd.EqualString(item.Obj.Network) {
			cpy, err := util.Clone(item.Obj)
			if err != nil {
				return nil, gcpmeta.NewInternalServerError("%v: failed to clone service connection: %v", common.ErrLogical, err)
			}
			result = append(result, cpy)
		}
	}

	return result, nil
}

func (s *store) CreateServiceConnection(ctx context.Context, projectId, vpcId string, reservedIpRanges []string) (*servicenetworking.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	netNd := gcputil.NewGlobalNetworkName(projectId, vpcId)
	net, err := s.getNetworkNoLock(projectId, vpcId)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("network %s not found", netNd.String())
	}

	var reservedPeeringRanges []string
	for _, reservedIpRange := range reservedIpRanges {
		// reservedIpRange is the name of the address
		addrNd := gcputil.NewGlobalAddressName(projectId, reservedIpRange)
		addr, found := s.addresses.FindByName(addrNd)
		if !found {
			return nil, gcpmeta.NewBadRequestError("address %s not found", addrNd)
		}
		reservedPeeringRanges = append(reservedPeeringRanges, addr.GetSelfLink())
	}

	con := &servicenetworking.Connection{
		Network:               net.GetSelfLink(),
		Peering:               gcpclient.PsaPeeringName,
		ReservedPeeringRanges: reservedPeeringRanges,
		Service:               gcpclient.ServiceNetworkingServiceConnectionName,
	}

	// indexing connection by network name
	s.serviceConnections.Add(con, netNd)

	// don't know the exact operation name format, this will do
	opName := s.newLongRunningOperationName(projectId)
	op := &servicenetworking.Operation{
		Done: true,
		Name: opName.String(),
	}
	s.serviceNetworkingOperations.Add(op, opName)
	return op, nil
}

// Operations ======================================

func (s *store) GetServiceNetworkingOperation(ctx context.Context, operationName string) (*servicenetworking.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	if util.IsContextDone(ctx) {
		return nil, ctx.Err()
	}

	opName, err := gcputil.ParseNameDetail(operationName)
	if err != nil {
		return nil, gcpmeta.NewBadRequestError("invalid operation name")
	}

	op, found := s.serviceNetworkingOperations.FindByName(opName)
	if !found {
		return nil, gcpmeta.NewNotFoundError("operation %s not found", operationName)
	}

	return op, nil
}
