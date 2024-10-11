package mock

import (
	"context"
	"fmt"
	"sync"

	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	iprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	nfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	redisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/redisinstance/client"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
	scopeclient "github.com/kyma-project/cloud-manager/pkg/kcp/scope/client"
)

var _ Server = &server{}

func New() Server {

	return &server{
		scopeStore: &scopeStore{},
		accounts:   map[string]*accountRegionStore{},
	}
}

type server struct {
	m sync.Mutex

	*scopeStore

	accounts map[string]*accountRegionStore
}

func (s *server) ScopeGardenProvider() awsclient.GardenClientProvider[scopeclient.AwsStsClient] {
	return func(ctx context.Context, region, key, secret string) (scopeclient.AwsStsClient, error) {
		return s, nil
	}
}

func (s *server) IpRangeSkrProvider() awsclient.SkrClientProvider[iprangeclient.Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (iprangeclient.Client, error) {
		return s.getAccountRegionContext(account, region), nil
	}
}

func (s *server) NfsInstanceSkrProvider() awsclient.SkrClientProvider[nfsinstanceclient.Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (nfsinstanceclient.Client, error) {
		return s.getAccountRegionContext(account, region), nil
	}
}

func (s *server) VpcPeeringSkrProvider() awsclient.SkrClientProvider[vpcpeeringclient.Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (vpcpeeringclient.Client, error) {
		return s.getAccountRegionContext(account, region), nil
	}
}

func (s *server) ElastiCacheProviderFake() awsclient.SkrClientProvider[redisinstanceclient.ElastiCacheClient] {
	return func(ctx context.Context, account, region, key, secret, role string) (redisinstanceclient.ElastiCacheClient, error) {
		return s.getAccountRegionContext(account, region), nil
	}
}

func (s *server) MockConfigs(account, region string) AccountRegion {
	return s.getAccountRegionContext(account, region)
}

func (s *server) getAccountRegionContext(account, region string) *accountRegionStore {
	s.m.Lock()
	defer s.m.Unlock()

	key := fmt.Sprintf("%s:%s", account, region)
	acc, ok := s.accounts[key]
	if !ok {
		acc = newAccountRegionStore(account, region)
		s.accounts[key] = acc
	}

	return acc
}
