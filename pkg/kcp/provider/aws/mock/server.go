package mock

import (
	"context"
	awsclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/aws/client"
	iprangeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/aws/iprange/client"
	nfsinstanceclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/aws/nfsinstance/client"
	scopeclient "github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/scope/client"
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
