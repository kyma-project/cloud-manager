package rediscluster

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/go-logr/logr"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azureredisclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/rediscluster/client"
	redisclustertypes "github.com/kyma-project/cloud-manager/pkg/kcp/rediscluster/types"
)

type State struct {
	redisclustertypes.State

	client         azureredisclient.Client
	provider       azureclient.ClientProvider[azureredisclient.Client]
	clientId       string
	clientSecret   string
	subscriptionId string
	tenantId       string

	resourceGroupName   string
	privateEndPoint     *armnetwork.PrivateEndpoint
	privateDnsZoneGroup *armnetwork.PrivateDNSZoneGroup
	azureRedisCluster   *armredis.ResourceInfo
}

type StateFactory interface {
	NewState(ctx context.Context, state redisclustertypes.State, logger logr.Logger) (*State, error)
}

type stateFactory struct {
	skrProvider azureclient.ClientProvider[azureredisclient.Client]
}

func NewStateFactory(skrProvider azureclient.ClientProvider[azureredisclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, redisclusterState redisclustertypes.State, logger logr.Logger) (*State, error) {

	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := redisclusterState.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := redisclusterState.Scope().Spec.Scope.Azure.TenantId

	c, err := f.skrProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)

	if err != nil {
		return nil, err
	}

	return newState(redisclusterState, c, f.skrProvider, clientId, clientSecret, subscriptionId, tenantId), nil
}

func newState(state redisclustertypes.State,
	client azureredisclient.Client,
	provider azureclient.ClientProvider[azureredisclient.Client],
	clientId string,
	clientSecret string,
	subscriptionId string,
	tenantId string) *State {
	return &State{
		State:          state,
		client:         client,
		provider:       provider,
		clientId:       clientId,
		clientSecret:   clientSecret,
		subscriptionId: subscriptionId,
		tenantId:       tenantId,

		resourceGroupName: azurecommon.AzureCloudManagerResourceGroupName(state.Scope().Spec.Scope.Azure.VpcNetwork),
	}
}
