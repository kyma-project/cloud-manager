package commonAction

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
)

type State interface {
	composed.State

	ObjAsObjWithConditionsAndState() composed.ObjWithConditionsAndState

	Subscription() *cloudcontrolv1beta1.Subscription
}

// factory ========================================================================

type StateFactory interface {
	NewState(base composed.State) State
}

func NewStateFactory() StateFactory {
	return &stateFactory{}
}

type stateFactory struct{}

func (f *stateFactory) NewState(base composed.State) State {
	return &stateImpl{
		State: base,
	}
}

// state ========================================================================

type stateImpl struct {
	composed.State

	subscription *cloudcontrolv1beta1.Subscription
}

func (s *stateImpl) ObjAsObjWithConditionsAndState() composed.ObjWithConditionsAndState {
	return s.Obj().(composed.ObjWithConditionsAndState)
}

func (s *stateImpl) Subscription() *cloudcontrolv1beta1.Subscription {
	return s.subscription
}
