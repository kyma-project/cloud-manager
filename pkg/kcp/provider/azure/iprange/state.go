package iprange

import (
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/kcp/provider/azure/client"
)

type State struct {
	focal.State

	client client.TbdAzureClient
}

type StateFactory interface {
	NewState(focalState focal.State) *State
}

type stateFactory struct {
	client client.TbdAzureClient
}

func (f *stateFactory) NewState(focalState focal.State) *State {
	return NewState(focalState, f.client)
}

func NewStateFactory(client client.TbdAzureClient) StateFactory {
	return &stateFactory{client: client}
}

func NewState(focalState focal.State, client client.TbdAzureClient) *State {
	return &State{
		State:  focalState,
		client: client,
	}
}
