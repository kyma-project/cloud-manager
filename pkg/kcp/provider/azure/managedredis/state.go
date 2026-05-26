package managedredis

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redisenterprise/armredisenterprise/v3"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	kcpcommonaction "github.com/kyma-project/cloud-manager/pkg/kcp/commonAction"
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
	kcpcommonaction.State

	scope *cloudcontrolv1beta1.Scope

	client            azuremanagedredisclient.Client
	resourceGroupName string

	managedRedis         *armredisenterprise.Cluster
	managedRedisDatabase *armredisenterprise.Database
	privateEndpoint      *armnetwork.PrivateEndpoint
	privateDnsZoneGroup  *armnetwork.PrivateDNSZoneGroup
}

func (s *State) ObjAsAzureManagedRedis() *cloudcontrolv1beta1.AzureManagedRedis {
	return s.Obj().(*cloudcontrolv1beta1.AzureManagedRedis)
}

func (s *State) Scope() *cloudcontrolv1beta1.Scope {
	return s.scope
}

func (s *State) SetScope(scope *cloudcontrolv1beta1.Scope) {
	s.scope = scope
}

func (s *State) PrivateDNSZoneName() string {
	if azureconfig.AzureConfig.ClientOptions.Cloud == "AzureChina" {
		return PrivateDNSZoneChina
	}
	return PrivateDNSZone
}

// initAzureClient finalizes State by creating the Azure client and setting the
// resource group name once Scope is loaded. It runs after loadScope.
func initAzureClient(clientProvider azureclient.ClientProvider[azuremanagedredisclient.Client]) composed.Action {
	return func(ctx context.Context, st composed.State) (error, context.Context) {
		state := st.(*State)
		if state.client != nil {
			return nil, ctx
		}

		clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
		clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
		subscriptionId := state.Scope().Spec.Scope.Azure.SubscriptionId
		tenantId := state.Scope().Spec.Scope.Azure.TenantId

		c, err := clientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
		if err != nil {
			return composed.LogErrorAndReturn(err, "Error creating Azure ManagedRedis client", composed.StopWithRequeue, ctx)
		}

		state.client = c
		state.resourceGroupName = azurecommon.AzureCloudManagerResourceGroupName(state.Scope().Spec.Scope.Azure.VpcNetwork)
		return nil, ctx
	}
}
