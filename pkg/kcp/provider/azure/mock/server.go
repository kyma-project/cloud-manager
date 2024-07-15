package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	provider "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	"k8s.io/utils/ptr"
)

var _ Server = &server{}

func New() Server {
	return &server{
		stores: map[string]*storeSubscriptionContext{},
	}
}

type server struct {
	stores map[string]*storeSubscriptionContext
}

func (s *server) VpcPeeringSkrProvider() provider.SkrClientProvider[client.Client] {
	return func(ctx context.Context, clientId, clientSecret, subscription, tenant string) (client.Client, error) {

		return s.getStoreSubscriptionContext(subscription), nil
	}
}

func (s *server) RedisClientProvider() provider.SkrClientProvider[azureredisinstanceclient.Client] {
	return func(ctx context.Context, clientId, clientSecret, subscription, tenant string) (azureredisinstanceclient.Client, error) {

		return nil, nil
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
