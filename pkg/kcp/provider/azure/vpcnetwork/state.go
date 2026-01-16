package vpcnetwork

import (
	"context"
	"fmt"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azurevpcnetworkclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcnetwork/client"
	vpcnetworktypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork/types"
)

func NewStateFactory(azureClientProvider azureclient.ClientProvider[azurevpcnetworkclient.Client]) StateFactory {
	return &stateFactory{
		azureClientProvider: azureClientProvider,
	}
}

type StateFactory interface {
	NewState(ctx context.Context, baseState vpcnetworktypes.State) (context.Context, composed.State, error)
}

type stateFactory struct {
	azureClientProvider azureclient.ClientProvider[azurevpcnetworkclient.Client]
}

func (f *stateFactory) NewState(ctx context.Context, baseState vpcnetworktypes.State) (context.Context, composed.State, error) {
	clientId := azureconfig.AzureConfig.DefaultCreds.ClientId
	clientSecret := azureconfig.AzureConfig.DefaultCreds.ClientSecret
	if baseState.Subscription().Status.Provider != cloudcontrolv1beta1.ProviderAzure {
		return ctx, nil, fmt.Errorf("subscription for VpcNetwork must be of provider Azure, but subscription %q is of provider %q", baseState.Subscription().Name, baseState.Subscription().Status.Provider)
	}
	tenantId := baseState.Subscription().Status.SubscriptionInfo.Azure.TenantId
	subscriptionId := baseState.Subscription().Status.SubscriptionInfo.Azure.SubscriptionId

	c, err := f.azureClientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return ctx, nil, fmt.Errorf("error creating azure client: %w", err)
	}

	logger := composed.LoggerFromCtx(ctx).
		WithValues(
			"azureSubscriptionId", subscriptionId,
			"azureTenantId", tenantId,
		)
	ctx = composed.LoggerIntoCtx(ctx, logger)

	return ctx, newState(baseState, c), nil
}

func newState(baseState vpcnetworktypes.State, azureClient azurevpcnetworkclient.Client) *State {
	return &State{
		State:       baseState,
		azureClient: azureClient,
	}
}

type State struct {
	vpcnetworktypes.State

	azureClient azurevpcnetworkclient.Client
}
