package mock

import (
	"context"
	"fmt"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/exposedData/client"
	azureiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/iprange/client"
	azurenetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/network/client"
	azureredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/rediscluster/client"
	azureredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	azurevpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
	"sync"
)

var _ Server = &server{}

func New() Server {
	return &server{
		subscriptions: map[string]*tenantSubscriptionStore{},
	}
}

var _ Server = &server{}

type server struct {
	m             sync.Mutex
	subscriptions map[string]*tenantSubscriptionStore
}

func (s *server) StorageProvider() azureclient.ClientProvider[azurerwxvolumebackupclient.Client] {
	return func(_ context.Context, _, _, subscription, tenant string, auxiliaryTenants ...string) (azurerwxvolumebackupclient.Client, error) {
		return s.getTenantStoreSubscriptionContext(subscription, tenant), nil
	}
}

func (s *server) IpRangeProvider() azureclient.ClientProvider[azureiprangeclient.Client] {
	return func(_ context.Context, _, _, subscription, tenant string, auxiliaryTenants ...string) (azureiprangeclient.Client, error) {
		return s.getTenantStoreSubscriptionContext(subscription, tenant), nil
	}
}

func (s *server) VpcPeeringProvider() azureclient.ClientProvider[azurevpcpeeringclient.Client] {
	return func(_ context.Context, _, _, subscription, tenant string, auxiliaryTenants ...string) (azurevpcpeeringclient.Client, error) {
		return s.getTenantStoreSubscriptionContext(subscription, tenant), nil
	}
}

func (s *server) RedisClientProvider() azureclient.ClientProvider[azureredisinstanceclient.Client] {
	return func(ctx context.Context, _, _, subscription, tenant string, auxiliaryTenants ...string) (azureredisinstanceclient.Client, error) {
		return s.getTenantStoreSubscriptionContext(subscription, tenant), nil
	}
}

func (s *server) RedisClusterClientProvider() azureclient.ClientProvider[azureredisclusterclient.Client] {
	return func(ctx context.Context, _, _, subscription, tenant string, auxiliaryTenants ...string) (azureredisclusterclient.Client, error) {
		return s.getTenantStoreSubscriptionContext(subscription, tenant), nil
	}
}

func (s *server) NetworkProvider() azureclient.ClientProvider[azurenetworkclient.Client] {
	return func(_ context.Context, _, _, subscription, tenant string, auxiliaryTenants ...string) (azurenetworkclient.Client, error) {
		return s.getTenantStoreSubscriptionContext(subscription, tenant), nil
	}
}

func (s *server) ExposeDataProvider() azureclient.ClientProvider[azureexposeddataclient.Client] {
	return func(ctx context.Context, clientId, clientSecret, subscriptionId, tenantId string, _ ...string) (azureexposeddataclient.Client, error) {
		return s.getTenantStoreSubscriptionContext(subscriptionId, tenantId), nil
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
