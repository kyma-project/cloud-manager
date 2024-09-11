package network

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	networktypes "github.com/kyma-project/cloud-manager/pkg/kcp/network/types"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azurenetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/network/client"
)

type State struct {
	networktypes.State

	azureClient azurenetworkclient.Client

	location          string
	tags              map[string]string
	resourceGroupName string
	cidr              string

	resourceGroup *armresources.ResourceGroup

	network *armnetwork.VirtualNetwork
}

type StateFactory interface {
	NewState(ctx context.Context, networkState networktypes.State) (*State, error)
}

type stateFactory struct {
	azureProvider azureclient.SkrClientProvider[azurenetworkclient.Client]
}

func NewStateFactory(azureProvider azureclient.SkrClientProvider[azurenetworkclient.Client]) StateFactory {
	return &stateFactory{
		azureProvider: azureProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, networkState networktypes.State) (*State, error) {
	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := networkState.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := networkState.Scope().Spec.Scope.Azure.TenantId

	c, err := f.azureProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return nil, err
	}

	return newState(networkState, c), nil
}

func newState(
	networkState networktypes.State,
	azureClient azurenetworkclient.Client,
) *State {
	return &State{
		State:       networkState,
		azureClient: azureClient,
	}
}
