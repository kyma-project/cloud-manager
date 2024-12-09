package mock

import (
	"context"
	"errors"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/smithy-go"
	"github.com/elliotchance/pie/v2"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/utils/ptr"
	"sync"
)

type VpcPeeringConfig interface {
	SetVpcPeeringConnectionStatusCode(requesterVpcId, accepterVpcId string, code ec2types.VpcPeeringConnectionStateReasonCode) error
	InitiateVpcPeeringConnection(connectionId, requesterVpcId, accepterVpcId string)
	SetVpcPeeringConnectionError(connectionId string, err error)
}

type vpcPeeringEntry struct {
	peering ec2types.VpcPeeringConnection
}
type vpcPeeringStore struct {
	m     sync.Mutex
	items []*vpcPeeringEntry

	errorMap map[string]error
}

func newVpcPeeringStore() *vpcPeeringStore {
	return &vpcPeeringStore{
		errorMap: make(map[string]error),
	}
}

func (s *vpcPeeringStore) CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string, tags []ec2types.Tag) (*ec2types.VpcPeeringConnection, error) {
	s.m.Lock()
	defer s.m.Unlock()

	var peering *ec2types.VpcPeeringConnection
	var err error

	peering, err = s.findVpcPeeringConnection(vpcId, remoteVpcId)

	if err != nil {
		peering = &ec2types.VpcPeeringConnection{
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
		}

		item := &vpcPeeringEntry{
			peering: *peering,
		}

		s.items = append(s.items, item)

	}

	peeringCopy, err := util.JsonClone(peering)

	if err != nil {
		return nil, err
	}

	return peeringCopy, nil
}

func (s *vpcPeeringStore) findVpcPeeringConnection(vpcId, remoteVpcId *string) (*ec2types.VpcPeeringConnection, error) {
	for _, x := range s.items {
		if ptr.Equal(x.peering.RequesterVpcInfo.VpcId, vpcId) ||
			ptr.Equal(x.peering.AccepterVpcInfo.VpcId, remoteVpcId) {
			return &x.peering, nil
		}
	}

	return nil, &smithy.GenericAPIError{
		Code:    "404",
		Message: fmt.Sprintf("vpc peering connection between %s and %s does not exist", ptr.Deref(vpcId, ""), ptr.Deref(remoteVpcId, "")),
	}
}

func (s *vpcPeeringStore) DescribeVpcPeeringConnection(ctx context.Context, vpcPeeringConnectionId string) (*ec2types.VpcPeeringConnection, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if err, ok := s.errorMap[vpcPeeringConnectionId]; ok && err != nil {
		return nil, err
	}

	peering, err := s.getVpcPeeringConnection(vpcPeeringConnectionId)

	if err != nil {
		return nil, err
	}

	peeringCopy, err := util.JsonClone(peering)

	if err != nil {
		return nil, err
	}

	return peeringCopy, nil
}

func (s *vpcPeeringStore) getVpcPeeringConnection(vpcPeeringConnectionId string) (*ec2types.VpcPeeringConnection, error) {
	for _, x := range s.items {
		if ptr.Equal(x.peering.VpcPeeringConnectionId, ptr.To(vpcPeeringConnectionId)) {
			return &x.peering, nil
		}
	}

	return nil, &smithy.GenericAPIError{
		Code:    "404",
		Message: fmt.Sprintf("vpc peering connection %s does not exist", vpcPeeringConnectionId),
	}
}

func (s *vpcPeeringStore) DescribeVpcPeeringConnections(ctx context.Context) ([]ec2types.VpcPeeringConnection, error) {
	s.m.Lock()
	defer s.m.Unlock()

	peerings := pie.Map(s.items, func(e *vpcPeeringEntry) ec2types.VpcPeeringConnection {
		return e.peering
	})

	peeringsCopy, err := util.JsonClone(peerings)

	if err != nil {
		return nil, err
	}

	return peeringsCopy, nil
}

func (s *vpcPeeringStore) AcceptVpcPeeringConnection(ctx context.Context, connectionId *string) (*ec2types.VpcPeeringConnection, error) {
	s.m.Lock()
	defer s.m.Unlock()

	if err, ok := s.errorMap[*connectionId]; ok {
		return nil, err
	}

	peering, err := s.getVpcPeeringConnection(*connectionId)

	if err != nil {
		return nil, err
	}

	peeringCopy, err := util.JsonClone(peering)

	if err != nil {
		return nil, err
	}

	return peeringCopy, nil
}

func (s *vpcPeeringStore) DeleteVpcPeeringConnection(ctx context.Context, connectionId *string) error {
	s.m.Lock()
	defer s.m.Unlock()

	if err, ok := s.errorMap[*connectionId]; ok && err != nil {
		return err
	}

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

func (s *vpcPeeringStore) SetVpcPeeringConnectionStatusCode(requesterVpcId, accepterVpcId string, code ec2types.VpcPeeringConnectionStateReasonCode) error {
	s.m.Lock()
	defer s.m.Unlock()

	item, err := s.findVpcPeeringConnection(ptr.To(requesterVpcId), ptr.To(accepterVpcId))

	if err != nil {
		return err
	}

	item.Status.Code = code

	return nil
}

func (s *vpcPeeringStore) InitiateVpcPeeringConnection(connectionId, requesterVpcId, accepterVpcId string) {
	s.m.Lock()
	defer s.m.Unlock()

	item := &vpcPeeringEntry{
		peering: ec2types.VpcPeeringConnection{
			VpcPeeringConnectionId: ptr.To(connectionId),
			RequesterVpcInfo: &ec2types.VpcPeeringConnectionVpcInfo{
				VpcId: ptr.To(requesterVpcId),
			},
			AccepterVpcInfo: &ec2types.VpcPeeringConnectionVpcInfo{
				VpcId: ptr.To(accepterVpcId),
			},
			Status: &ec2types.VpcPeeringConnectionStateReason{
				Code:    ec2types.VpcPeeringConnectionStateReasonCodeInitiatingRequest,
				Message: nil,
			},
		},
	}

	s.items = append(s.items, item)
}

func (s *vpcPeeringStore) SetVpcPeeringConnectionError(connectionId string, err error) {
	s.errorMap[connectionId] = err
}
