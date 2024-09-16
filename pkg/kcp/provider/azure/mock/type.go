package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	provider "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	networkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/network/client"
	redisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
)

// ResourceClient not implemented yet in the reconcilers so fixed here as the future reference, since some methods needed in other places
type ResourceClient interface {
	networkclient.ResourceGroupClient
}

// NetworkClient not implemented yet in the reconcilers so fixed here as the future reference, since some methods needed in peering
type NetworkClient interface {
	networkclient.NetworkClient
}

type VpcPeeringClient interface {
	vpcpeeringclient.Client
}

type RedisInstanceClient interface {
	redisinstanceclient.Client
}

type Clients interface {
	ResourceClient
	NetworkClient
	VpcPeeringClient
	RedisInstanceClient
}

type Providers interface {
	VpcPeeringSkrProvider() provider.ClientProvider[vpcpeeringclient.Client]
	RedisClientProvider() provider.ClientProvider[redisinstanceclient.Client]
	NetworkProvider() provider.ClientProvider[networkclient.Client]
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
