package focal

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-resources/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-resources/components/lib/composed"
)

type State interface {
	composed.State
	Scope() *cloudresourcesv1beta1.Scope
	SetScope(*cloudresourcesv1beta1.Scope)
	CommonObj() CommonObject
}

type StateFactory interface {
	NewState(base composed.State) State
}

func NewStateFactory() StateFactory {
	return &stateFactory{}
}

type stateFactory struct{}

func (f *stateFactory) NewState(base composed.State) State {
	return &state{
		State: base,
	}
}

// ========================================================================

type state struct {
	composed.State

	scope *cloudresourcesv1beta1.Scope
}

func (s *state) Scope() *cloudresourcesv1beta1.Scope {
	return s.scope
}

func (s *state) SetScope(scope *cloudresourcesv1beta1.Scope) {
	s.scope = scope
}

func (s *state) CommonObj() CommonObject {
	return s.Obj().(CommonObject)
}
