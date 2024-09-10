package mock

import (
	"context"
	"errors"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"k8s.io/utils/ptr"
	"sync"
)

type VpcPeeringConfig interface {
	SetVpcPeeringConnectionActive(ctx context.Context, vpcId, remoteVpcId *string)
}

type vpcPeeringEntry struct {
	peering ec2types.VpcPeeringConnection
}
type vpcPeeringStore struct {
	m     sync.Mutex
	items []*vpcPeeringEntry
}

func (s *vpcPeeringStore) CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string, tags []ec2types.Tag) (*ec2types.VpcPeeringConnection, error) {
	s.m.Lock()
	defer s.m.Unlock()

	item := &vpcPeeringEntry{
		peering: ec2types.VpcPeeringConnection{
			VpcPeeringConnectionId: ptr.To("pcx-" + uuid.NewString()[:8]),
			RequesterVpcInfo: &ec2types.VpcPeeringConnectionVpcInfo{
				VpcId: vpcId,
			},
			AccepterVpcInfo: &ec2types.VpcPeeringConnectionVpcInfo{
				VpcId:   remoteVpcId,
				Region:  remoteRegion,
				OwnerId: remoteAccountId,
			},
			Status: &ec2types.VpcPeeringConnectionStateReason{
				Code:    ec2types.VpcPeeringConnectionStateReasonCodeInitiatingRequest,
				Message: nil,
			},
		},
	}

	s.items = append(s.items, item)

	return &item.peering, nil
}

func (s *vpcPeeringStore) DescribeVpcPeeringConnections(ctx context.Context) ([]ec2types.VpcPeeringConnection, error) {
	s.m.Lock()
	defer s.m.Unlock()

	return pie.Map(s.items, func(e *vpcPeeringEntry) ec2types.VpcPeeringConnection {
		return e.peering
	}), nil
}

func (s *vpcPeeringStore) AcceptVpcPeeringConnection(ctx context.Context, connectionId *string) (*ec2types.VpcPeeringConnection, error) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, x := range s.items {
		if ptr.Equal(x.peering.VpcPeeringConnectionId, connectionId) {
			return &x.peering, nil
		}
	}

	return nil, fmt.Errorf("an error occurred (InvalidVpcPeeringConnectionID.NotFound) when calling the AcceptVpcPeeringConnection operation: The vpcPeeringConnection ID %s' does not exist", *connectionId)
}

func (s *vpcPeeringStore) DeleteVpcPeeringConnection(ctx context.Context, connectionId *string) error {
	s.m.Lock()
	defer s.m.Unlock()

	deleted := false
	s.items = pie.Filter(s.items, func(x *vpcPeeringEntry) bool {
		deleted = ptr.Equal(x.peering.VpcPeeringConnectionId, connectionId)
		return !deleted
	})

	if !deleted {
		return errors.New("peering connection not found")
	}

	return nil
}

func (s *vpcPeeringStore) SetVpcPeeringConnectionActive(ctx context.Context, vpcId, remoteVpcId *string) {
	s.m.Lock()
	defer s.m.Unlock()

	for _, x := range s.items {
		if ptr.Equal(x.peering.AccepterVpcInfo.VpcId, remoteVpcId) &&
			ptr.Equal(x.peering.RequesterVpcInfo.VpcId, vpcId) {
			x.peering.Status.Code = ec2types.VpcPeeringConnectionStateReasonCodeActive
			break
		}
	}
}
