package iprange

import "github.com/kyma-project/cloud-resources/components/kcp/pkg/common/actions/scope"

type State struct {
	scope.State
}

func newState(scopeState scope.State) *State {
	return &State{State: scopeState}
}
