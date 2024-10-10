package nuke

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StateFactory interface {
	NewState(req ctrl.Request) *State
}

func NewStateFactory(
	baseStateFactory composed.StateFactory,
	activeSkrCollection skrruntime.ActiveSkrCollection,
) StateFactory {
	return &stateFactory{
		baseStateFactory:    baseStateFactory,
		activeSkrCollection: activeSkrCollection,
	}
}

type stateFactory struct {
	baseStateFactory    composed.StateFactory
	activeSkrCollection skrruntime.ActiveSkrCollection
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	baseState := f.baseStateFactory.NewState(req.NamespacedName, &cloudcontrolv1beta1.Nuke{})

	return newState(baseState, f.activeSkrCollection)
}

func newState(
	baseState composed.State,
	activeSkrCollection skrruntime.ActiveSkrCollection,
) *State {
	return &State{
		State:               baseState,
		activeSkrCollection: activeSkrCollection,
	}
}

type ResourceKindState struct {
	Kind    string
	List    client.ObjectList
	Objects []focal.CommonObject
}

type State struct {
	composed.State
	activeSkrCollection skrruntime.ActiveSkrCollection

	Resources []*ResourceKindState
}

func (s *State) ObjAsNuke() *cloudcontrolv1beta1.Nuke {
	return s.Obj().(*cloudcontrolv1beta1.Nuke)
}

func (s *State) ObjectExists(kind, name string) bool {
	for _, res := range s.Resources {
		if res.Kind == kind {
			for _, obj := range res.Objects {
				if obj.GetName() == name {
					return true
				}
			}
		}
	}
	return false
}
