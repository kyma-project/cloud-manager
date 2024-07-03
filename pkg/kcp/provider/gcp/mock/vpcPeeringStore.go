package mock

import (
	compute "cloud.google.com/go/compute/apiv1"
	pb "cloud.google.com/go/compute/apiv1/computepb"
	"context"
	"fmt"
	"k8s.io/utils/ptr"
	"sync"
)

type vpcPeeringEntry struct {
	peering *pb.NetworkPeering
}
type vpcPeeringStore struct {
	m     sync.Mutex
	items map[string]*vpcPeeringEntry
}

func getFullNetworkUrl(project, vpc string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", project, vpc)
}

func (s *vpcPeeringStore) CreateRemoteVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, importCustomRoutes bool, kymaProject string, kymaVpc string) (*compute.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	remoteNetwork := getFullNetworkUrl(remoteProject, remoteVpc)
	kymaNetwork := getFullNetworkUrl(kymaProject, kymaVpc)

	_, peeringExists := s.items[remoteNetwork]
	if peeringExists {
		return new(compute.Operation), nil
	}

	state := pb.NetworkPeering_ACTIVE.String()

	item := &vpcPeeringEntry{
		peering: &pb.NetworkPeering{
			Name:                 &remotePeeringName,
			Network:              &kymaNetwork,
			ImportCustomRoutes:   &importCustomRoutes,
			ExchangeSubnetRoutes: ptr.To(true),
		},
	}
	item.peering.State = &state
	s.items[remoteNetwork] = item

	return new(compute.Operation), nil
}

func (s *vpcPeeringStore) CreateKymaVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, importCustomRoutes bool, kymaProject string, kymaVpc string) (*compute.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()

	remoteNetwork := getFullNetworkUrl(remoteProject, remoteVpc)
	kymaNetwork := getFullNetworkUrl(kymaProject, kymaVpc)

	_, peeringExists := s.items[kymaNetwork]
	if peeringExists {
		return new(compute.Operation), nil
	}

	state := pb.NetworkPeering_ACTIVE.String()

	item := &vpcPeeringEntry{
		peering: &pb.NetworkPeering{
			Name:                 &remotePeeringName,
			Network:              &remoteNetwork,
			ImportCustomRoutes:   &importCustomRoutes,
			ExchangeSubnetRoutes: ptr.To(true),
		},
	}
	item.peering.State = &state

	s.items[kymaNetwork] = item

	return new(compute.Operation), nil
}

func (s *vpcPeeringStore) CheckRemoteNetworkTags(context context.Context, remoteVpc string, remoteProject string, desiredTag string) (bool, error) {
	s.m.Lock()
	defer s.m.Unlock()

	return true, nil
}

func (s *vpcPeeringStore) GetVpcPeering(ctx context.Context, remotePeeringName string, project string, vpc string) (*pb.NetworkPeering, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if s.items == nil {
		s.items = make(map[string]*vpcPeeringEntry)
	}

	network := getFullNetworkUrl(project, vpc)

	_, peeringExists := s.items[network]
	if !peeringExists {
		return nil, nil
	}

	return s.items[network].peering, nil
}

func (s *vpcPeeringStore) DeleteVpcPeering(ctx context.Context, remotePeeringName string, kymaProject string, kymaVpc string) (*compute.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()

	kymaNetwork := getFullNetworkUrl(kymaProject, kymaVpc)

	if s.items[kymaNetwork] == nil {
		return nil, nil
	}
	delete(s.items, kymaNetwork)
	return new(compute.Operation), nil
}
