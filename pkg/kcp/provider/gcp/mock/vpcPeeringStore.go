package mock

import (
	pb "cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/elliotchance/pie/v2"
	"sync"
)

type vpcPeeringEntry struct {
	peering pb.NetworkPeering
}
type vpcPeeringStore struct {
	m     sync.Mutex
	items []*vpcPeeringEntry
}

func (s *vpcPeeringStore) CreateVpcPeeringConnection(ctx context.Context, name *string, remoteVpc *string, remoteProject *string, importCustomRoutes *bool, kymaProject *string, kymaVpc *string) (pb.NetworkPeering, error) {
	s.m.Lock()
	defer s.m.Unlock()

	item := &vpcPeeringEntry{
		peering: pb.NetworkPeering{
			Name:                 name,
			Network:              remoteVpc,
			ImportCustomRoutes:   importCustomRoutes,
			ExchangeSubnetRoutes: to.Ptr(true),
		},
	}

	s.items = append(s.items, item)

	return item.peering, nil
}

func (s *vpcPeeringStore) DescribeVpcPeeringConnections(ctx context.Context) ([]pb.NetworkPeering, error) {
	s.m.Lock()
	defer s.m.Unlock()

	return pie.Map(s.items, func(e *vpcPeeringEntry) pb.NetworkPeering {
		return e.peering
	}), nil
}
