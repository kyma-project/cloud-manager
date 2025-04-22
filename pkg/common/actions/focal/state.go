package focal

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

type State interface {
	composed.State
	Scope() *cloudcontrolv1beta1.Scope
	SetScope(*cloudcontrolv1beta1.Scope)
	ObjAsCommonObj() CommonObject

	setScopeOptional()
	isScopeOptional() bool
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

	scope *cloudcontrolv1beta1.Scope

	_optionalScope bool
}

func (s *state) Scope() *cloudcontrolv1beta1.Scope {
	return s.scope
}

func (s *state) SetScope(scope *cloudcontrolv1beta1.Scope) {
	s.scope = scope
}

func (s *state) ObjAsCommonObj() CommonObject {
	return s.Obj().(CommonObject)
}

func (s *state) isScopeOptional() bool {
	return s._optionalScope
}

func (s *state) setScopeOptional() {
	s._optionalScope = true
}
