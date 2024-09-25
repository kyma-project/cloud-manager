package vpcpeering

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/config"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	vpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	clientProvider azureclient.ClientProvider[vpcpeeringclient.Client]
	localClient    vpcpeeringclient.Client
	remoteClient   vpcpeeringclient.Client

	localNetwork    *cloudcontrolv1beta1.Network
	localNetworkId  *azureutil.NetworkResourceId
	remoteNetwork   *cloudcontrolv1beta1.Network
	remoteNetworkId *azureutil.NetworkResourceId

	localPeering  *armnetwork.VirtualNetworkPeering
	remotePeering *armnetwork.VirtualNetworkPeering
	remoteVpc     *armnetwork.VirtualNetwork
}

type StateFactory interface {
	NewState(ctx context.Context, state vpcpeeringtypes.State, logger logr.Logger) (*State, error)
}

type stateFactory struct {
	clientProvider azureclient.ClientProvider[vpcpeeringclient.Client]
}

func NewStateFactory(skrProvider azureclient.ClientProvider[vpcpeeringclient.Client]) StateFactory {
	return &stateFactory{
		clientProvider: skrProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, logger logr.Logger) (*State, error) {
	clientId := azureconfig.AzureConfig.PeeringCreds.ClientId
	clientSecret := azureconfig.AzureConfig.PeeringCreds.ClientSecret
	subscriptionId := vpcPeeringState.Scope().Spec.Scope.Azure.SubscriptionId
	tenantId := vpcPeeringState.Scope().Spec.Scope.Azure.TenantId

	c, err := f.clientProvider(ctx, clientId, clientSecret, subscriptionId, tenantId)
	if err != nil {
		return nil, fmt.Errorf("error creating local azure client: %w", err)
	}

	return newState(vpcPeeringState, c, f.clientProvider), nil
}

func newState(state vpcpeeringtypes.State,
	client vpcpeeringclient.Client,
	clientProvider azureclient.ClientProvider[vpcpeeringclient.Client],
) *State {
	return &State{
		State:          state,
		localClient:    client,
		clientProvider: clientProvider,
	}
}
