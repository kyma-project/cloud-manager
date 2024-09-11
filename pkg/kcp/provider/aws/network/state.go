package network

import (
	"context"
	"errors"
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

func (f *stateFactory) NewState(ctx context.Context, networkState networktypes.State) (*State, error) {
	return nil, errors.New("AWS Network not implemented")
}
