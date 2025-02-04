package rediscluster

import (
	"context"

	"github.com/kyma-project/cloud-manager/pkg/common/abstractions"

	"github.com/kyma-project/cloud-manager/pkg/kcp/rediscluster/types"
)

type State struct {
	types.State
}

type StateFactory interface {
	NewState(ctx context.Context, redisClusterState types.State) (*State, error)
}

type stateFactory struct {
	env abstractions.Environment
}

func NewStateFactory(env abstractions.Environment) StateFactory {
	return &stateFactory{
		env: env,
	}
}

func (statefactory *stateFactory) NewState(ctx context.Context, redisClusterState types.State) (*State, error) {

	return newState(redisClusterState), nil
}

func newState(redisClusterState types.State) *State {
	return &State{
		State: redisClusterState,
	}
}
