package dnszone

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vnetlink/dnszone/client"
)

type State struct {
	focal.State

	clientProvider azureclient.ClientProvider[client.Client]
	remoteClient   client.Client

	vnetLink               *armprivatedns.VirtualNetworkLink
	privateDnzZone         *armprivatedns.PrivateZone
	remotePrivateDnsZoneId azureutil.ResourceDetails
}

type StateFactory interface {
	NewState(ctx context.Context, focalState focal.State) (*State, error)
}

type stateFactory struct {
	clientProvider azureclient.ClientProvider[client.Client]
}

func NewStateFactory(clientProvider azureclient.ClientProvider[client.Client]) StateFactory {
	return &stateFactory{
		clientProvider: clientProvider,
	}
}

func (f *stateFactory) NewState(ctx context.Context, focalState focal.State) (*State, error) {
	return newState(focalState, f.clientProvider), nil
}

func newState(focalState focal.State, provider azureclient.ClientProvider[client.Client]) *State {
	return &State{
		State:          focalState,
		clientProvider: provider,
	}
}

func (s *State) ObjAsAzureVNetLink() *v1beta1.AzureVNetLink {
	return s.Obj().(*v1beta1.AzureVNetLink)
}
