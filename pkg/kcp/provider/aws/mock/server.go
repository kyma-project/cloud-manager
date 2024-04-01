package mock

import (
	"context"
	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	nfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
)

var _ Server = &server{}

func New() Server {
	return &server{
		vpcStore:   &vpcStore{},
		nfsStore:   &nfsStore{},
		scopeStore: &scopeStore{},
	}
}

type server struct {
	*vpcStore
	*nfsStore
	*scopeStore
}

func (s *server) ScopeGardenProvider() awsclient.GardenClientProvider[scopeclient.AwsStsClient] {
	return func(ctx context.Context, region, key, secret string) (scopeclient.AwsStsClient, error) {
		return s, nil
	}
}

func (s *server) IpRangeSkrProvider() awsclient.SkrClientProvider[iprangeclient.Client] {
	return func(ctx context.Context, region, key, secret, role string) (iprangeclient.Client, error) {
		return s, nil
	}
}

func (s *server) NfsInstanceSkrProvider() awsclient.SkrClientProvider[nfsinstanceclient.Client] {
	return func(ctx context.Context, region, key, secret, role string) (nfsinstanceclient.Client, error) {
		return s, nil
	}
}

func (s *server) VpcPeeringSkrProvider() awsclient.SkrClientProvider[vpcpeeringclient.Client] {
	return func(ctx context.Context, region, key, secret, role string) (vpcpeeringclient.Client, error) {
		return s, nil
	}
}
