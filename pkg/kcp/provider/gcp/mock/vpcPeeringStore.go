package mock

import (
	compute "cloud.google.com/go/compute/apiv1"
	pb "cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"github.com/elliotchance/pie/v2"
	"k8s.io/utils/ptr"
	"sync"
)

type vpcPeeringEntry struct {
	peering *pb.NetworkPeering
}
type vpcPeeringStore struct {
	m     sync.Mutex
	items []*vpcPeeringEntry
}

func (s *vpcPeeringStore) CreateVpcPeering(ctx context.Context, name *string, remoteVpc *string, remoteProject *string, importCustomRoutes *bool, kymaProject *string, kymaVpc *string) (*pb.NetworkPeering, error) {
	s.m.Lock()
	defer s.m.Unlock()

	item := &vpcPeeringEntry{
		peering: &pb.NetworkPeering{
			Name:                 name,
			Network:              remoteVpc,
			ImportCustomRoutes:   importCustomRoutes,
			ExchangeSubnetRoutes: ptr.To(true),
		},
	}

	s.items = append(s.items, item)

	return item.peering, nil
}

func (s *vpcPeeringStore) DeleteVpcPeering(ctx context.Context, name *string, kymaProject *string, kymaVpc *string) (*compute.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	s.items = pie.Filter(s.items, func(vpe *vpcPeeringEntry) bool {
		return !(vpe.peering.Name == name && *vpe.peering.Network == "https://www.googleapis.com/compute/v1/projects/"+*kymaProject+"/global/networks/"+*kymaVpc)
	})
	return nil, nil
}

func (s *vpcPeeringStore) DescribeVpcPeeringConnections(ctx context.Context) ([]*pb.NetworkPeering, error) {
	s.m.Lock()
	defer s.m.Unlock()

	return pie.Map(s.items, func(e *vpcPeeringEntry) *pb.NetworkPeering {
		return e.peering
	}), nil
}
