package iprange

import (
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
)

type State struct {
	focal.State
}

type StateFactory interface {
	NewState(focalState focal.State) *State
}

type stateFactory struct {
}

func (f *stateFactory) NewState(focalState focal.State) *State {
	return NewState(focalState)
}

func NewStateFactory() StateFactory {
	return &stateFactory{}
}

func NewState(focalState focal.State) *State {
	return &State{
		State: focalState,
	}
}
