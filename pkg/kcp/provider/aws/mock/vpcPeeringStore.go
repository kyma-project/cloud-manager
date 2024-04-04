package mock

import (
	"context"
	"fmt"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"k8s.io/utils/pointer"
	"sync"
)

type VpcPeeringConfig interface {
	AddVpcPeering()
}

type vpcPeeringStore struct {
	m     sync.Mutex
	items []string
}

func (s *vpcPeeringStore) CreateVpcPeeringConnection(ctx context.Context, vpcId, remoteVpcId, remoteRegion, remoteAccountId *string) (*ec2types.VpcPeeringConnection, error) {
	s.m.Lock()
	defer s.m.Unlock()

	c := ec2types.VpcPeeringConnection{
		VpcPeeringConnectionId: pointer.String(fmt.Sprintf("%s->%s",
			pointer.StringDeref(vpcId, ""),
			pointer.StringDeref(remoteVpcId, ""))),
		RequesterVpcInfo: &ec2types.VpcPeeringConnectionVpcInfo{
			VpcId: vpcId,
		},
		AccepterVpcInfo: &ec2types.VpcPeeringConnectionVpcInfo{
			VpcId:   remoteVpcId,
			Region:  remoteRegion,
			OwnerId: remoteAccountId,
		},
	}

	s.items = append(s.items, pointer.StringDeref(c.VpcPeeringConnectionId, ""))

	return &c, nil
}
