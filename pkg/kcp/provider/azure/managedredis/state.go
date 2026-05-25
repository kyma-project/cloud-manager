package managedredis

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"
	"github.com/go-logr/logr"

	managedredistypes "github.com/kyma-project/cloud-manager/pkg/kcp/managedredis/types"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azuremanagedredisclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/managedredis/client"
)

const (
	PrivateEndpointGroupID = "redisEnterprise"
	PrivateDNSZone         = "privatelink.redis.azure.net"
	PrivateDNSZoneChina    = "privatelink.redis.cache.chinacloudapi.cn"
	RedisPort              = int32(10000)
	DefaultDatabaseName    = "default"
)

type State struct {
	managedredistypes.State

	client            azuremanagedredisclient.Client
	resourceGroupName string

	managedRedis         *armredisenterprise.Cluster
	managedRedisDatabase *armredisenterprise.Database
	privateEndpoint      *armnetwork.PrivateEndpoint
	privateDnsZoneGroup  *armnetwork.PrivateDNSZoneGroup
}

func (s *State) PrivateDNSZoneName() string {
	if azureconfig.AzureConfig.ClientOptions.Cloud == "AzureChina" {
		return PrivateDNSZoneChina
	}
	return PrivateDNSZone
}

type StateFactory interface {
	NewState(ctx context.Context, managedRedisState managedredistypes.State, logger logr.Logger) (*State, error)
}

type stateFactory struct {
	clientProvider azureclient.ClientProvider[azuremanagedredisclient.Client]
}

func NewStateFactory(clientProvider azureclient.ClientProvider[azuremanagedredisclient.Client]) StateFactory {
	return &stateFactory{
		clientProvider: clientProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, managedRedisState managedredistypes.State, _ logr.Logger) (*State, error) {
	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := managedRedisState.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := managedRedisState.Scope().Spec.Scope.Azure.TenantId

	c, err := f.clientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return nil, err
	}

	return &State{
		State:             managedRedisState,
		client:            c,
		resourceGroupName: azurecommon.AzureCloudManagerResourceGroupName(managedRedisState.Scope().Spec.Scope.Azure.VpcNetwork),
	}, nil
}
