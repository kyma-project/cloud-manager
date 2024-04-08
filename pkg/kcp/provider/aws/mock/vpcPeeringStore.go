package mock

import (
	"context"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"k8s.io/utils/pointer"
	"sync"
)

type VpcPeeringConfig interface {
	AddVpcPeering(PeeringConnectionId, vpcId, remoteVpcId, remoteRegion, remoteAccountId string) *ec2types.VpcPeeringConnection
}

type vpcPeeringEntry struct {
	peering ec2types.VpcPeeringConnection
}
type vpcPeeringStore struct {
	m     sync.Mutex
	items []*vpcPeeringEntry
}

func (s *vpcPeeringStore) AddVpcPeering(id, vpcId, remoteVpcId, remoteRegion, remoteAccountId string) *ec2types.VpcPeeringConnection {
	item := &vpcPeeringEntry{
		peering: ec2types.VpcPeeringConnection{
			VpcPeeringConnectionId: pointer.String(id),
			RequesterVpcInfo: &ec2types.VpcPeeringConnectionVpcInfo{
				VpcId: pointer.String(vpcId),
			},
			AccepterVpcInfo: &ec2types.VpcPeeringConnectionVpcInfo{
				VpcId:   pointer.String(remoteVpcId),
				Region:  pointer.String(remoteRegion),
				OwnerId: pointer.String(remoteAccountId),
			},
		},
	}

	s.items = append(s.items, item)
	return &item.peering
}

func (s *vpcPeeringStore) CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string) (*ec2types.VpcPeeringConnection, error) {
	s.m.Lock()
	defer s.m.Unlock()

	idx := pie.FindFirstUsing(s.items, func(e *vpcPeeringEntry) bool {
		return pointer.StringEqual(e.peering.RequesterVpcInfo.VpcId, vpcId) &&
			pointer.StringEqual(e.peering.AccepterVpcInfo.VpcId, remoteVpcId)
	})

	return &s.items[idx].peering, nil
}
