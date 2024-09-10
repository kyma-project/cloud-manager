package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	provider "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	redisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
)

// ResourceClient not implemented yet in the reconcilers so fixed here as the future reference, since some methods needed in other places
type ResourceClient interface {
	GetResourceGroup(ctx context.Context, name string) (*armresources.ResourceGroup, error)
	CreateResourceGroup(ctx context.Context, name string, location string, tags map[string]string) (*armresources.ResourceGroup, error)
}

// NetworkClient not implemented yet in the reconcilers so fixed here as the future reference, since some methods needed in peering
type NetworkClient interface {
	CreateNetwork(ctx context.Context, resourceGroupName, virtualNetworkName, location, addressSpace string, tags map[string]string) (*armnetwork.VirtualNetwork, error)
	GetNetwork(ctx context.Context, resourceGroupName, virtualNetworkName string) (*armnetwork.VirtualNetwork, error)
}

type VpcPeeringClient interface {
	vpcpeeringclient.Client
}

type RedisInstanceClient interface {
	redisinstanceclient.Client
}

type Clients interface {
	NetworkClient
	VpcPeeringClient
	RedisInstanceClient
}

type Providers interface {
	VpcPeeringSkrProvider() provider.SkrClientProvider[vpcpeeringclient.Client]

	RedisClientProvider() provider.SkrClientProvider[redisinstanceclient.Client]
}

type NetworkConfig interface {
	SetPeeringStateConnected(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) error
}

type RedisConfig interface {
	AzureRemoveRedisInstance(ctx context.Context, resourceGroupName, redisInstanceName string) error
	AzureSetRedisInstanceState(ctx context.Context, resourceGroupName, redisInstanceName string, state armredis.ProvisioningState) error
}

type Configs interface {
	NetworkConfig
	RedisConfig
}

type TenantSubscription interface {
	Clients
	Configs
}

type Server interface {
	Providers

	MockConfigs(subscription, tenant string) TenantSubscription
}
