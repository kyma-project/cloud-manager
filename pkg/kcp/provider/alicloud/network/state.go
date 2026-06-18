package network

import (
	"context"
	networktypes "github.com/kyma-project/cloud-manager/pkg/kcp/network/types"
)

type State struct {
	networktypes.State
}

type StateFactory interface {
	NewState(ctx context.Context, networkState networktypes.State) (*State, error)
}

func NewStateFactory() StateFactory {
	return &stateFactory{}
}

type stateFactory struct{}

func (f *stateFactory) NewState(_ context.Context, networkState networktypes.State) (*State, error) {
	return &State{State: networkState}, nil
}
