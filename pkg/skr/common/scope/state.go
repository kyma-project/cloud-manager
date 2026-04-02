package scope

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type State interface {
	composed.State
	KcpCluster() composed.StateCluster
	KymaRef() klog.ObjectRef
	Scope() *cloudcontrolv1beta1.Scope
	ObjWithConditionsAndState() composed.ObjWithConditionsAndState
	SetScope(scope *cloudcontrolv1beta1.Scope)
}

type StateFactory interface {
	NewState(ctx context.Context, namespacedName types.NamespacedName, baseState composed.State) (State, error)
}

func NewStateFactory(kcpCluster composed.StateCluster, scopeProvider scopeprovider.ScopeProvider) StateFactory {
	return &stateFactory{
		kcpCluster:    kcpCluster,
		scopeProvider: scopeProvider,
	}
}

type stateFactory struct {
	kcpCluster    composed.StateCluster
	scopeProvider scopeprovider.ScopeProvider
}

func (f *stateFactory) NewState(ctx context.Context, namespacedName types.NamespacedName, baseState composed.State) (State, error) {
	kymaRef, err := f.scopeProvider.GetScope(ctx, namespacedName)
	if err != nil {
		return nil, err
	}
	return &myState{
		State:      baseState,
		kcpCluster: f.kcpCluster,
		kymaRef:    kymaRef,
	}, nil
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

func (s *myState) SetScope(scope *cloudcontrolv1beta1.Scope) {
	s.scope = scope
}
