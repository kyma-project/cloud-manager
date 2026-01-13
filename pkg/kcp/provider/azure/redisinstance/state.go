package redisinstance

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/go-logr/logr"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azureredisinstanceclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	redisinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
)

type State struct {
	redisinstancetypes.State

	client         azureredisinstanceclient.Client
	provider       azureclient.ClientProvider[azureredisinstanceclient.Client]
	clientId       string
	clientSecret   string
	subscriptionId string
	tenantId       string

	resourceGroupName   string
	privateEndPoint     *armnetwork.PrivateEndpoint
	privateDnsZoneGroup *armnetwork.PrivateDNSZoneGroup
	azureRedisInstance  *armredis.ResourceInfo
}

type StateFactory interface {
	NewState(ctx context.Context, state redisinstancetypes.State, logger logr.Logger) (*State, error)
}

type stateFactory struct {
	skrProvider azureclient.ClientProvider[azureredisinstanceclient.Client]
}

func NewStateFactory(skrProvider azureclient.ClientProvider[azureredisinstanceclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, redisinstanceState redisinstancetypes.State, logger logr.Logger) (*State, error) {

	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := redisinstanceState.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := redisinstanceState.Scope().Spec.Scope.Azure.TenantId

	c, err := f.skrProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)

	if err != nil {
		return nil, err
	}

	return newState(redisinstanceState, c, f.skrProvider, clientId, clientSecret, subscriptionId, tenantId), nil
}

func newState(state redisinstancetypes.State,
	client azureredisinstanceclient.Client,
	provider azureclient.ClientProvider[azureredisinstanceclient.Client],
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

// GetProvisionedMachineType returns the provisioned machine type from the Azure Redis Instance
func (s *State) GetProvisionedMachineType() string {
	if s.azureRedisInstance == nil || s.azureRedisInstance.Properties == nil || s.azureRedisInstance.Properties.SKU == nil {
		return ""
	}
	return fmt.Sprintf("%s%d", *s.azureRedisInstance.Properties.SKU.Family, *s.azureRedisInstance.Properties.SKU.Capacity)
}

// GetProvisionedMemorySizeGb returns the provisioned memory size in GB from the Azure Redis Instance
func (s *State) GetProvisionedMemorySizeGb() int32 {
	if s.azureRedisInstance == nil || s.azureRedisInstance.Properties == nil || s.azureRedisInstance.Properties.SKU == nil {
		return 0
	}
	return *s.azureRedisInstance.Properties.SKU.Capacity
}

// GetProvisionedReplicaCount returns the provisioned replica count from the Azure Redis Instance
func (s *State) GetProvisionedReplicaCount() int32 {
	if s.azureRedisInstance == nil || s.azureRedisInstance.Properties == nil || s.azureRedisInstance.Properties.ShardCount == nil {
		return 0
	}
	return *s.azureRedisInstance.Properties.ShardCount
}
