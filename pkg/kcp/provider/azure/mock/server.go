package mock

import (
	"context"
	provider "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
)

var _ Server = &server{}

func New() Server {
	return &server{
		vpcPeeringStore: &vpcPeeringStore{},
	}
}

type server struct {
	*vpcPeeringStore
}

func (s *server) VpcPeeringSkrProvider() provider.SkrClientProvider[client.Client] {
	return func(ctx context.Context, region, key, secret, role string) (client.Client, error) {
		return s, nil
	}
}
