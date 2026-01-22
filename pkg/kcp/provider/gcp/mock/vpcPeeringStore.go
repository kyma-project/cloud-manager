package mock

import (
	"context"
	"fmt"
	"sync"

	pb "cloud.google.com/go/compute/apiv1/computepb"
	"k8s.io/utils/ptr"
)

type vpcPeeringEntry struct {
	peering *pb.NetworkPeering
}
type vpcPeeringStore struct {
	m          sync.Mutex
	items      map[string]*vpcPeeringEntry
	operations map[string]*pb.Operation
	errorMap   map[string]error
	tags       map[string][]string
}

type VpcPeeringMockClientUtils interface {
	GetMockVpcPeering(project string, vpc string) *pb.NetworkPeering
	SetMockVpcPeeringLifeCycleState(project string, vpc string, state pb.NetworkPeering_State)
	SetMockVpcPeeringError(project string, vpc string, err error)
	SetMockVpcPeeringTags(project string, vpc string, tags []string)
}

func getFullNetworkUrl(project, vpc string) string {
	return fmt.Sprintf("https://www.googleapis.com/compute/v1/projects/%s/global/networks/%s", project, vpc)
}

func (s *vpcPeeringStore) CreateRemoteVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, customRoutes bool, kymaProject string, kymaVpc string) (*pb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()
	remoteNetwork := getFullNetworkUrl(remoteProject, remoteVpc)
	kymaNetwork := getFullNetworkUrl(kymaProject, kymaVpc)

	exportCustomRoutes := false
	importCustomRoutes := false
	if customRoutes {
		exportCustomRoutes = true
	}

	if s.operations == nil {
		s.operations = make(map[string]*pb.Operation)
	}

	if s.items == nil {
		s.items = make(map[string]*vpcPeeringEntry)
	}

	if s.operations[remoteVpc] == nil {
		operationpb := &pb.Operation{
			Status: ptr.To(pb.Operation_DONE),
			Name:   ptr.To(remoteVpc),
		}
		s.operations[remoteVpc] = operationpb
	}

	item := &vpcPeeringEntry{
		peering: &pb.NetworkPeering{
			Name:                 &remotePeeringName,
			Network:              &kymaNetwork,
			ImportCustomRoutes:   &importCustomRoutes,
			ExportCustomRoutes:   &exportCustomRoutes,
			ExchangeSubnetRoutes: ptr.To(true),
			State:                ptr.To(pb.NetworkPeering_INACTIVE.String()),
		},
	}
	s.items[remoteNetwork] = item

	return s.operations[remoteVpc], nil
}

func (s *vpcPeeringStore) CreateLocalVpcPeering(ctx context.Context, remotePeeringName string, remoteVpc string, remoteProject string, customRoutes bool, kymaProject string, kymaVpc string) (*pb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()

	remoteNetwork := getFullNetworkUrl(remoteProject, remoteVpc)
	kymaNetwork := getFullNetworkUrl(kymaProject, kymaVpc)

	exportCustomRoutes := false
	importCustomRoutes := customRoutes

	if s.operations == nil {
		s.operations = make(map[string]*pb.Operation)
	}

	if s.items == nil {
		s.items = make(map[string]*vpcPeeringEntry)
	}

	if s.operations[kymaVpc] == nil {
		operationpb := &pb.Operation{
			Status: ptr.To(pb.Operation_DONE),
			Name:   ptr.To(kymaVpc),
		}
		s.operations[kymaVpc] = operationpb
	}

	item := &vpcPeeringEntry{
		peering: &pb.NetworkPeering{
			Name:                 &remotePeeringName,
			Network:              &remoteNetwork,
			ImportCustomRoutes:   &importCustomRoutes,
			ExportCustomRoutes:   &exportCustomRoutes,
			ExchangeSubnetRoutes: ptr.To(true),
			State:                ptr.To(pb.NetworkPeering_INACTIVE.String()),
		},
	}

	s.items[kymaNetwork] = item

	return s.operations[kymaVpc], nil
}

func (s *vpcPeeringStore) GetRemoteNetworkTags(_ context.Context, vpc string, project string) ([]string, error) {
	s.m.Lock()
	defer s.m.Unlock()

	return s.tags[getFullNetworkUrl(project, vpc)], nil
}

func (s *vpcPeeringStore) GetVpcPeering(ctx context.Context, remotePeeringName string, project string, vpc string) (*pb.NetworkPeering, error) {
	s.m.Lock()
	defer s.m.Unlock()

	network := getFullNetworkUrl(project, vpc)

	if s.errorMap == nil {
		s.errorMap = make(map[string]error)
	}

	if err, errorExists := s.errorMap[network]; errorExists {
		return nil, err
	}

	if s.items == nil {
		s.items = make(map[string]*vpcPeeringEntry)
	}

	if s.tags == nil {
		s.tags = make(map[string][]string)
	}

	if s.operations == nil {
		s.operations = make(map[string]*pb.Operation)
	}

	_, peeringExists := s.items[network]
	if !peeringExists {
		return nil, nil
	}

	return s.items[network].peering, nil
}

func (s *vpcPeeringStore) DeleteVpcPeering(ctx context.Context, remotePeeringName string, kymaProject string, kymaVpc string) error {
	s.m.Lock()
	defer s.m.Unlock()

	kymaNetwork := getFullNetworkUrl(kymaProject, kymaVpc)

	if s.items[kymaNetwork] == nil {
		return nil
	}
	delete(s.items, kymaNetwork)
	return nil
}

func (s *vpcPeeringStore) GetOperation(ctx context.Context, project string, operationId string) (*pb.Operation, error) {
	s.m.Lock()
	defer s.m.Unlock()

	_, operationExists := s.operations[operationId]
	if !operationExists {
		return nil, fmt.Errorf("\"The resource 'projects/%s/global/operations/%s' was not found\",", project, operationId)
	}
	return s.operations[operationId], nil
}

// Fake client implementations to mimic google API calls
func (s *vpcPeeringStore) SetMockVpcPeeringLifeCycleState(project string, vpc string, state pb.NetworkPeering_State) {
	stateString := state.String()
	if s.items[getFullNetworkUrl(project, vpc)] != nil {
		s.items[getFullNetworkUrl(project, vpc)].peering.State = &stateString
	}
}

func (s *vpcPeeringStore) GetMockVpcPeering(project string, vpc string) *pb.NetworkPeering {
	_, peeringExists := s.items[getFullNetworkUrl(project, vpc)]
	if !peeringExists {
		return nil
	}
	return s.items[getFullNetworkUrl(project, vpc)].peering
}

func (s *vpcPeeringStore) SetMockVpcPeeringError(project string, vpc string, err error) {
	if s.errorMap == nil {
		s.errorMap = make(map[string]error)
	}
	s.errorMap[getFullNetworkUrl(project, vpc)] = err
}

func (s *vpcPeeringStore) SetMockVpcPeeringTags(project string, vpc string, tags []string) {
	if s.tags == nil {
		s.tags = make(map[string][]string)
	}
	s.tags[getFullNetworkUrl(project, vpc)] = tags
}
