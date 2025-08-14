package mock

import (
	"context"
	"fmt"
	"sync"

	awsexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/exposedData/client"
	subscriptionclient "github.com/kyma-project/cloud-manager/pkg/kcp/subscription/client"

	awsclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/client"
	awsiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/iprange/client"
	awsnfsinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/nfsinstance/client"
	awsvpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/vpcpeering/client"
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

func (s *server) SubscriptionGardenProvider() awsclient.GardenClientProvider[subscriptionclient.AwsStsClient] {
	return func(ctx context.Context, region, key, secret string) (subscriptionclient.AwsStsClient, error) {
		return s, nil
	}
}

func (s *server) IpRangeSkrProvider() awsclient.SkrClientProvider[awsiprangeclient.Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (awsiprangeclient.Client, error) {
		return s.getAccountRegionContext(account, region), nil
	}
}

func (s *server) NfsInstanceSkrProvider() awsclient.SkrClientProvider[awsnfsinstanceclient.Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (awsnfsinstanceclient.Client, error) {
		return s.getAccountRegionContext(account, region), nil
	}
}

func (s *server) VpcPeeringSkrProvider() awsclient.SkrClientProvider[awsvpcpeeringclient.Client] {
	return func(ctx context.Context, account, region, key, secret, role string) (awsvpcpeeringclient.Client, error) {
		return s.getAccountRegionContext(account, region), nil
	}
}

func (s *server) ElastiCacheProviderFake() awsclient.SkrClientProvider[awsclient.ElastiCacheClient] {
	return func(ctx context.Context, account, region, key, secret, role string) (awsclient.ElastiCacheClient, error) {
		return s.getAccountRegionContext(account, region), nil
	}
}

func (s *server) ExposedDataProvider() awsclient.SkrClientProvider[awsexposeddataclient.Client] {
	return func(_ context.Context, account, region, key, secret, role string) (awsexposeddataclient.Client, error) {
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
