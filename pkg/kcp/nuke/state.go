package nuke

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common/actions/focal"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	nuketypes "github.com/kyma-project/cloud-manager/pkg/kcp/nuke/types"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
)

type StateFactory interface {
	NewState(req ctrl.Request) *State
}

func NewStateFactory(
	baseStateFactory composed.StateFactory,
	focalStateFactory focal.StateFactory,
	activeSkrCollection skrruntime.ActiveSkrCollection,
) StateFactory {
	return &stateFactory{
		baseStateFactory:    baseStateFactory,
		focalStateFactory:   focalStateFactory,
		activeSkrCollection: activeSkrCollection,
	}
}

type stateFactory struct {
	baseStateFactory    composed.StateFactory
	focalStateFactory   focal.StateFactory
	activeSkrCollection skrruntime.ActiveSkrCollection
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	baseState := f.baseStateFactory.NewState(req.NamespacedName, &cloudcontrolv1beta1.Nuke{})
	focalState := f.focalStateFactory.NewState(baseState)
	return newState(focalState, f.activeSkrCollection)
}

func newState(
	focalState focal.State,
	activeSkrCollection skrruntime.ActiveSkrCollection,
) *State {
	return &State{
		State:               focalState,
		ActiveSkrCollection: activeSkrCollection,
	}
}

type State struct {
	focal.State
	ActiveSkrCollection skrruntime.ActiveSkrCollection
	Resources           []*nuketypes.ResourceKindState
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
