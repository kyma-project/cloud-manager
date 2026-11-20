package scope

import (
	"context"
	"fmt"

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
	ObjWithStatus() composed.ObjWithStatus
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
	_, isObjWithStatus := baseState.Obj().(composed.ObjWithStatus)
	_, isObjWithConditionsAndState := baseState.Obj().(composed.ObjWithConditionsAndState)
	if !isObjWithStatus && !isObjWithConditionsAndState {
		return nil, fmt.Errorf("object %T must preferably implement composed.ObjWithStatus, or eventually composed.ObjWithConditionsAndState", baseState.Obj())
	}
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

// ObjWithConditionsAndState returns object for legacy resources using composed.UpdateStatus.
// It is preferred to use newer API style with composed.ObjWithStatus and composed.StatusPatcherComposed.
// If this method returns nil, then use ObjWithStatus().
func (s *myState) ObjWithConditionsAndState() composed.ObjWithConditionsAndState {
	obj, ok := s.Obj().(composed.ObjWithConditionsAndState)
	if ok {
		return obj
	}
	return nil
}

// ObjWithStatus returns object for newer API style with composed.ObjWithStatus
// and composed.StatusPatcherComposed. This is a preferred usage pattern
// If this method returns nil then use ObjWithConditionsAndState().
func (s *myState) ObjWithStatus() composed.ObjWithStatus {
	obj, ok := s.Obj().(composed.ObjWithStatus)
	if ok {
		return obj
	}
	return nil
}

func (s *myState) SetScope(scope *cloudcontrolv1beta1.Scope) {
	s.scope = scope
}
