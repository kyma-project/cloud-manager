package exposedData

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azureexposeddataclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/exposedData/client"
	scopetypes "github.com/kyma-project/cloud-manager/pkg/kcp/scope/types"
)

func NewStateFactory(azureProvider azureclient.ClientProvider[azureexposeddataclient.Client]) StateFactory {
	return &stateFactory{
		azureProvider: azureProvider,
	}
}

type StateFactory interface {
	NewState(ctx context.Context, baseState scopetypes.State) (composed.State, error)
}

var _ StateFactory = &stateFactory{}

type stateFactory struct {
	azureProvider azureclient.ClientProvider[azureexposeddataclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, baseState scopetypes.State) (composed.State, error) {
	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	subscriptionId := baseState.ObjAsScope().Spec.Scope.Azure.SubscriptionId
	tenantId := baseState.ObjAsScope().Spec.Scope.Azure.TenantId

	c, err := f.azureProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return nil, fmt.Errorf("error creating azure client: %w", err)
	}
	return newState(baseState, c), nil
}

func newState(baseState scopetypes.State, azureClient azureexposeddataclient.Client) *State {
	return &State{
		State:       baseState,
		azureClient: azureClient,
	}
}

type State struct {
	scopetypes.State

	azureClient azureexposeddataclient.Client
}
