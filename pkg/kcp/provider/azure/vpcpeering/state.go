package vpcpeering

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/go-logr/logr"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	azurevpcpeeringclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vpcpeering/client"
	vpcpeeringtypes "github.com/kyma-project/cloud-manager/pkg/kcp/vpcpeering/types"
)

type State struct {
	vpcpeeringtypes.State

	clientProvider azureclient.ClientProvider[azurevpcpeeringclient.Client]
	// localClient is used for API calls in local subscription that does not require auxiliary tenants and therefore is
	// resilient to authentication errors when SPN does not exist in remote tenant.
	localClient azurevpcpeeringclient.Client
	// localPeeringClient is used for API calls in local subscription that requires auxiliary tenants. Client fails if
	// SPN does not exist in remote tenant.
	localPeeringClient azurevpcpeeringclient.Client
	// remoteClient is used for API calls in remote subscription and may require auxiliary tenants
	remoteClient azurevpcpeeringclient.Client

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
	clientProvider azureclient.ClientProvider[azurevpcpeeringclient.Client]
}

func NewStateFactory(skrProvider azureclient.ClientProvider[azurevpcpeeringclient.Client]) StateFactory {
	return &stateFactory{
		clientProvider: skrProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, vpcPeeringState vpcpeeringtypes.State, logger logr.Logger) (*State, error) {
	return newState(vpcPeeringState, f.clientProvider), nil
}

func newState(state vpcpeeringtypes.State,
	clientProvider azureclient.ClientProvider[azurevpcpeeringclient.Client],
) *State {
	return &State{
		State:          state,
		clientProvider: clientProvider,
	}
}
