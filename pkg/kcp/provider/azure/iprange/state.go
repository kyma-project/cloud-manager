package iprange

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	iprangetypes "github.com/kyma-project/cloud-manager/pkg/kcp/iprange/types"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azureiprange "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/iprange/client"
)

type State struct {
	iprangetypes.State

	azureClient azureiprange.Client

	resourceGroupName  string
	virtualNetworkName string
	securityGroupName  string
	subnetName         string

	subnet             *armnetwork.Subnet
	securityGroup      *armnetwork.SecurityGroup
	privateDnsZone     *armprivatedns.PrivateZone
	virtualNetworkLink *armprivatedns.VirtualNetworkLink
}

type StateFactory interface {
	NewState(ctx context.Context, baseState iprangetypes.State) (*State, error)
}

type stateFactory struct {
	azureProvider azureclient.ClientProvider[azureiprange.Client]
}

func (f *stateFactory) NewState(ctx context.Context, baseState iprangetypes.State) (*State, error) {
	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := baseState.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := baseState.Scope().Spec.Scope.Azure.TenantId

	c, err := f.azureProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return nil, fmt.Errorf("error creating azure client: %w", err)
	}
	return NewState(c, baseState), nil
}

func NewStateFactory(azureProvider azureclient.ClientProvider[azureiprange.Client]) StateFactory {
	return &stateFactory{
		azureProvider: azureProvider,
	}
}

func NewState(azureClient azureiprange.Client, baseState iprangetypes.State) *State {
	return &State{
		State:       baseState,
		azureClient: azureClient,
	}
}
