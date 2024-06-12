package vpcpeering

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/go-logr/logr"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	client         vpcpeeringclient.Client
	provider       azureclient.SkrClientProvider[vpcpeeringclient.Client]
	clientId       string
	clientSecret   string
	subscriptionId string
	tenantId       string

	peering       *armnetwork.VirtualNetworkPeering
	remotePeering *armnetwork.VirtualNetworkPeering
	remoteVpc     *armnetwork.VirtualNetwork
}

// goes to new.go
type StateFactory interface {
	NewState(ctx context.Context, state vpcpeeringtypes.State, logger logr.Logger) (*State, error)
}

type stateFactory struct {
	skrProvider azureclient.SkrClientProvider[vpcpeeringclient.Client]
}

func NewStateFactory(skrProvider azureclient.SkrClientProvider[vpcpeeringclient.Client]) StateFactory {
	return &stateFactory{
		skrProvider: skrProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, logger logr.Logger) (*State, error) {

	clientId := azureconfig.AzureConfig.ClientId
	clientSecret := azureconfig.AzureConfig.ClientSecret
	subscriptionId := vpcPeeringState.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := vpcPeeringState.Scope().Spec.Scope.Azure.TenantId

	c, err := f.skrProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)

	if err != nil {
		return nil, err
	}

	return newState(vpcPeeringState, c, f.skrProvider, clientId, clientSecret, subscriptionId, tenantId), nil
}

func newState(state vpcpeeringtypes.State,
	client vpcpeeringclient.Client,
	provider azureclient.SkrClientProvider[vpcpeeringclient.Client],
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
