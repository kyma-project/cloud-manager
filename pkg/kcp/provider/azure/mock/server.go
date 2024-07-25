package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	armRedis "github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	provider "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	"k8s.io/utils/ptr"
	"sync"
)

var _ Server = &server{}

func New() Server {
	return &server{
		stores: map[string]*storeSubscriptionContext{},
		redisCacheClientFake: redisCacheClientFake{
			mutex:               sync.Mutex{},
			redisResourceGroups: map[string]redisResourceGroup{},
		},
	}
}

type server struct {
	stores               map[string]*storeSubscriptionContext
	redisCacheClientFake redisCacheClientFake
}

func (s *server) VpcPeeringSkrProvider() provider.SkrClientProvider[client.Client] {
	return func(ctx context.Context, clientId, clientSecret, subscription, tenant string) (client.Client, error) {

		return s.getStoreSubscriptionContext(subscription), nil
	}
}

func (s *server) RedisClientProvider() provider.SkrClientProvider[azureredisinstanceclient.Client] {
	return func(ctx context.Context, clientId, clientSecret, subscription, tenant string) (azureredisinstanceclient.Client, error) {

		return &s.redisCacheClientFake, nil
	}
}

func (s *server) AddNetwork(subscription, resourceGroup, virtualNetworkName string, tags map[string]*string) {

	entry := &networkEntry{
		resourceGroup: resourceGroup,
		network: armnetwork.VirtualNetwork{
			Name: ptr.To(virtualNetworkName),
			Tags: tags,
		},
	}

	store := s.getStoreSubscriptionContext(subscription)

	store.networkStore.items = append(store.networkStore.items, entry)
}

func (s *server) getStoreSubscriptionContext(subscription string) *storeSubscriptionContext {

	if s.stores[subscription] == nil {
		s.stores[subscription] = &storeSubscriptionContext{
			peeringStore: &peeringStore{},
			networkStore: &networkStore{},
			subscription: subscription,
		}
	}

	return s.stores[subscription]
}

func (s *server) GetRedisCacheByResourceGroupName(resourceGroupName string) *armRedis.ResourceInfo {
	redisInstance, _ := s.redisCacheClientFake.GetRedisInstance(nil, resourceGroupName, "")

	return redisInstance
}

func (s *server) DeleteRedisCacheByResourceGroupName(resourceGroupName string) {
	delete(s.redisCacheClientFake.redisResourceGroups, resourceGroupName)
}
