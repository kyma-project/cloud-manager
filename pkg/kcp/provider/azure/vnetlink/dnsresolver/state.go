package dnsresolver

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/dnsresolver/armdnsresolver"
	"github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	azureclient "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/client"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/vnetlink/dnsresolver/client"
)

type State struct {
	focal.State

	clientProvider azureclient.ClientProvider[client.Client]
	remoteClient   client.Client

	vnetLink  *armdnsresolver.VirtualNetworkLink
	ruleset   *armdnsresolver.DNSForwardingRuleset
	rulesetId azureutil.ResourceDetails
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

func (s *stateFactory) NewState(_ context.Context, focalState focal.State) (*State, error) {
	return newState(focalState, s.clientProvider), nil
}

func newState(focalState focal.State, clientProvider azureclient.ClientProvider[client.Client]) *State {
	return &State{
		State:          focalState,
		clientProvider: clientProvider,
	}
}

func (s *State) ObjAsAzureVNetLink() *v1beta1.AzureVNetLink {
	return s.Obj().(*v1beta1.AzureVNetLink)
}
