package mock

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	dnsresolverclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vnetlink/dnsresolver/client"
	azurevnetlinkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vnetlink/dnszone/client"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/exposedData/client"
	azureiprangeclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/iprange/client"
	azurenetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/network/client"
	azureredisclusterclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/rediscluster/client"
	azureredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	azurevpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	azurerwxpvclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxpv/client"
	azurerwxvolumebackupclient "github.com/kyma-project/cloud-manager/pkg/skr/azurerwxvolumebackup/client"
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

type DnsResolverVNetLinkClient interface {
	azureclient.DnsResolverVNetLinkClient
}

type VpcPeeringClient interface {
	azureclient.VirtualNetworkPeeringClient
}

type RedisInstanceClient interface {
	azureclient.RedisClient
}

type NatGatewayClient interface {
	azureclient.NatGatewayClient
}

type PublicIpAddressesClient interface {
	azureclient.PublicIPAddressesClient
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
	NatGatewayClient
	PublicIpAddressesClient
}

type Providers interface {
	VpcPeeringProvider() azureclient.ClientProvider[azurevpcpeeringclient.Client]
	IpRangeProvider() azureclient.ClientProvider[azureiprangeclient.Client]
	RedisClientProvider() azureclient.ClientProvider[azureredisinstanceclient.Client]
	RedisClusterClientProvider() azureclient.ClientProvider[azureredisclusterclient.Client]
	NetworkProvider() azureclient.ClientProvider[azurenetworkclient.Client]
	StorageProvider() azureclient.ClientProvider[azurerwxvolumebackupclient.Client]
	ExposeDataProvider() azureclient.ClientProvider[azureexposeddataclient.Client]
	RwxPvProvider() azureclient.ClientProvider[azurerwxpvclient.Client]
	DnsZoneVNetLinkProvider() azureclient.ClientProvider[azurevnetlinkclient.Client]
	DnsResolverVNetLinkProvider() azureclient.ClientProvider[dnsresolverclient.Client]
}

type NetworkConfig interface {
	SetPeeringConnectedFullInSync(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string) error
	SetPeeringError(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string, err error)
	SetPeeringSyncLevel(ctx context.Context, resourceGroup, virtualNetworkName, virtualNetworkPeeringName string, peeringLevel armnetwork.VirtualNetworkPeeringLevel) error
	SetNetworkAddressSpace(ctx context.Context, resourceGroup, virtualNetworkName, addressSpace string) error
	AddRemoteSubscription(ctx context.Context, remoteSubscription *TenantSubscription)
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
	TenantId() string
	SubscriptionId() string
}

type Server interface {
	Providers

	MockConfigs(subscription, tenant string) TenantSubscription
}
