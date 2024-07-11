package redisinstance

import (
	"context"
	"github.com/go-logr/logr"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azureredisclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/redisinstance/client"
	redisinstancetypes "github.com/kyma-project/cloud-manager/pkg/kcp/redisinstance/types"
)

type State struct {
	redisinstancetypes.State

	client         azureredisclient.MemorystoreClient
	provider       azureclient.SkrClientProvider[azureredisclient.MemorystoreClient]
	clientId       string
	clientSecret   string
	subscriptionId string
	tenantId       string
}

type StateFactory interface {
	NewState(ctx context.Context, state redisinstancetypes.State, logger logr.Logger) (*State, error)
}

type stateFactory struct {
	skrProvider azureclient.SkrClientProvider[azureredisclient.MemorystoreClient]
}

func NewStateFactory(skrProvider azureclient.SkrClientProvider[azureredisclient.MemorystoreClient]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, redisinstanceState redisinstancetypes.State, logger logr.Logger) (*State, error) {

	clientId := azureconfig.AzureConfig.ClientId
	clientSecret := azureconfig.AzureConfig.ClientSecret
	subscriptionId := redisinstanceState.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := redisinstanceState.Scope().Spec.Scope.Azure.TenantId

	c, err := f.skrProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)

	if err != nil {
		return nil, err
	}

	return newState(redisinstanceState, c, f.skrProvider, clientId, clientSecret, subscriptionId, tenantId), nil
}

func newState(state redisinstancetypes.State,
	client azureredisclient.MemorystoreClient,
	provider azureclient.SkrClientProvider[azureredisclient.MemorystoreClient],
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
	}
}
