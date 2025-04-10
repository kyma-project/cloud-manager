package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/exposedData/client"
	azureiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/iprange/client"
	networkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/network/client"
	redisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/rediscluster/client"
	redisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	skrazureclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
)

type ResourceGroupsClient interface {
	azureclient.ResourceGroupClient
}

type NetworkClient interface {
	azureclient.NetworkClient
}

type SubnetsClient interface {
	azureclient.SubnetsClient
}

type SecurityGroupsClient interface {
	azureclient.SecurityGroupsClient
}

type VirtualNetworkLinkClient interface {
	azureclient.VirtualNetworkLinkClient
}

type PrivateEndPointsClient interface {
	azureclient.PrivateEndPointsClient
}

type PrivateDnsZoneGroupClient interface {
	azureclient.PrivateDnsZoneGroupClient
}

type PrivateDnsZoneClient interface {
	azureclient.PrivateDnsZoneClient
}

type VpcPeeringClient interface {
	azureclient.VirtualNetworkPeeringClient
}

type RedisInstanceClient interface {
	azureclient.RedisClient
}

type Clients interface {
	ResourceGroupsClient
	NetworkClient
	SecurityGroupsClient
	SubnetsClient
	VpcPeeringClient
	RedisInstanceClient
	VirtualNetworkLinkClient
	PrivateDnsZoneClient
	PrivateEndPointsClient
	PrivateDnsZoneGroupClient
}

type Providers interface {
	VpcPeeringProvider() azureclient.ClientProvider[vpcpeeringclient.Client]
	IpRangeProvider() azureclient.ClientProvider[azureiprangeclient.Client]
	RedisClientProvider() azureclient.ClientProvider[redisinstanceclient.Client]
	RedisClusterClientProvider() azureclient.ClientProvider[redisclusterclient.Client]
	NetworkProvider() azureclient.ClientProvider[networkclient.Client]
	StorageProvider() azureclient.ClientProvider[skrazureclient.Client]
	ExposeDataProvider() azureclient.ClientProvider[azureexposeddataclient.Client]
}

type NetworkConfig interface {
	SetPeeringStateConnected(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) error
	SetPeeringError(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string, err error)
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
