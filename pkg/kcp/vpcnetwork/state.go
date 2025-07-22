package vpcnetwork

import (
	"github.com/kyma-project/cloud-manager/pkg/composed"
	ctrl "sigs.k8s.io/controller-runtime"
)

type StateFactory interface {
	NewState(req ctrl.Request) *State
}

func NewStateFactory(
	baseStateFactory composed.StateFactory,
) StateFactory {
	return &stateFactory{
		baseStateFactory: baseStateFactory,
	}
}

type stateFactory struct {
	baseStateFactory composed.StateFactory
}

func (s *stateFactory) NewState(req ctrl.Request) *State {
	return &State{}
}

// State ==================================

type State struct {
	composed.State
}
