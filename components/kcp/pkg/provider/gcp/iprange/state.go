package iprange

import (
	"context"

	"github.com/kyma-project/cloud-manager/components/kcp/pkg/iprange/types"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/provider/gcp/client"
)

type State struct {
	types.State

	client client.TbdGcpClient
}

type StateFactory interface {
	NewState(ctx context.Context, ipRangeState types.State) (*State, error)
}

type stateFactory struct {
	client client.TbdGcpClient
}

func (f *stateFactory) NewState(ctx context.Context, ipRangeState types.State) (*State, error) {
	return NewState(ipRangeState, f.client), nil
}

func NewStateFactory(client client.TbdGcpClient) StateFactory {
	return &stateFactory{client: client}
}

func NewState(ipRangeState types.State, client client.TbdGcpClient) *State {
	return &State{
		State:  ipRangeState,
		client: client,
	}
}
