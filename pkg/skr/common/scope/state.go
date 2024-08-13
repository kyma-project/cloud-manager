package scope

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/klog/v2"
)

type State interface {
	composed.State
	KcpCluster() composed.StateCluster
	KymaRef() klog.ObjectRef
	Scope() *cloudcontrolv1beta1.Scope
	ObjWithConditionsAndState() composed.ObjWithConditionsAndState
	setScope(scope *cloudcontrolv1beta1.Scope)
}

type StateFactory interface {
	NewState(baseState composed.State) State
}

func NewStateFactory(kcpCluster composed.StateCluster, kymaRef klog.ObjectRef) StateFactory {
	return &stateFactory{
		kcpCluster: kcpCluster,
		kymaRef:    kymaRef,
	}
}

type stateFactory struct {
	kcpCluster composed.StateCluster
	kymaRef    klog.ObjectRef
}

func (f *stateFactory) NewState(baseState composed.State) State {
	return &myState{
		State:      baseState,
		kcpCluster: f.kcpCluster,
		kymaRef:    f.kymaRef,
	}
}

type myState struct {
	composed.State
	kcpCluster composed.StateCluster
	kymaRef    klog.ObjectRef

	scope *cloudcontrolv1beta1.Scope
}

func (s *myState) KcpCluster() composed.StateCluster {
	return s.kcpCluster
}

func (s *myState) KymaRef() klog.ObjectRef {
	return s.kymaRef
}

func (s *myState) Scope() *cloudcontrolv1beta1.Scope {
	return s.scope
}

func (s *myState) ObjWithConditionsAndState() composed.ObjWithConditionsAndState {
	return s.Obj().(composed.ObjWithConditionsAndState)
}

func (s *myState) setScope(scope *cloudcontrolv1beta1.Scope) {
	s.scope = scope
}
