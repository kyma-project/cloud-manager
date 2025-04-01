package kyma

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"github.com/kyma-project/cloud-manager/pkg/util"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	ctrl "sigs.k8s.io/controller-runtime"
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
	baseState := f.baseStateFactory.NewState(req.NamespacedName, util.NewKymaUnstructured())

	return newState(
		baseState,
		f.activeSkrCollection,
	)
}

func newState(baseState composed.State, activeSkrCollection skrruntime.ActiveSkrCollection) *State {
	return &State{
		State:               baseState,
		activeSkrCollection: activeSkrCollection,
	}
}

type State struct {
	composed.State

	activeSkrCollection skrruntime.ActiveSkrCollection

	// scope potentially can be nil in case when it does not exist
	scope *cloudcontrolv1beta1.Scope

	moduleState  util.KymaModuleState
	moduleInSpec bool
	skrActive    bool
}

func (s *State) ObjAsKyma() *unstructured.Unstructured {
	return s.Obj().(*unstructured.Unstructured)
}
