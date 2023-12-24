package iprange

import (
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-resources/components/kcp/pkg/provider/gcp/client"
)

type State struct {
	focal.State

	client client.TbdGcpClient
}

type StateFactory interface {
	NewState(focalState focal.State) *State
}

type stateFactory struct {
	client client.TbdGcpClient
}

func (f *stateFactory) NewState(focalState focal.State) *State {
	return NewState(focalState, f.client)
}

func NewStateFactory(client client.TbdGcpClient) StateFactory {
	return &stateFactory{client: client}
}

func NewState(focalState focal.State, client client.TbdGcpClient) *State {
	return &State{
		State:  focalState,
		client: client,
	}
}
