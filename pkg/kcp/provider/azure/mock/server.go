package mock

import (
	"context"
	"fmt"
	provider "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	networkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/network/client"
	redisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	"sync"
)

var _ Server = &server{}

func New() Server {
	return &server{
		subscriptions: map[string]*tenantSubscriptionStore{},
	}
}

type server struct {
	m             sync.Mutex
	subscriptions map[string]*tenantSubscriptionStore
}

func (s *server) VpcPeeringSkrProvider() provider.SkrClientProvider[vpcpeeringclient.Client] {
	return func(_ context.Context, _, _, subscription, tenant string) (vpcpeeringclient.Client, error) {
		return s.getTenantStoreSubscriptionContext(subscription, tenant), nil
	}
}

func (s *server) RedisClientProvider() provider.SkrClientProvider[redisinstanceclient.Client] {
	return func(ctx context.Context, _, _, subscription, tenant string) (redisinstanceclient.Client, error) {
		return s.getTenantStoreSubscriptionContext(subscription, tenant), nil
	}
}

func (s *server) NetworkProvider() provider.SkrClientProvider[networkclient.Client] {
	return func(_ context.Context, _, _, subscription, tenant string) (networkclient.Client, error) {
		return s.getTenantStoreSubscriptionContext(subscription, tenant), nil
	}
}

// MockConfigs returns all configs for the given subscription and tenant, and
// should be used in tests for perform resource changes beside the official API
// for things that normally cloud providers do in the background
func (s *server) MockConfigs(subscription, tenant string) TenantSubscription {
	return s.getTenantStoreSubscriptionContext(subscription, tenant)
}

func (s *server) getTenantStoreSubscriptionContext(subscription, tenant string) *tenantSubscriptionStore {
	s.m.Lock()
	defer s.m.Unlock()

	key := fmt.Sprintf("%s:%s", tenant, subscription)
	sub, ok := s.subscriptions[key]
	if !ok {
		sub = newTenantSubscriptionStore(tenant, subscription)
		s.subscriptions[key] = sub
	}

	return sub
}
