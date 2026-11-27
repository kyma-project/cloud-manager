package cloudresources

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	commonscope "github.com/kyma-project/cloud-manager/pkg/skr/common/scope"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	commonscope.State

	Provider *cloudcontrolv1beta1.ProviderType

	IgnoreWatchErrors func(bool)
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	scopeStateFactory commonscope.StateFactory,
	provider *cloudcontrolv1beta1.ProviderType,
	ignoreWatchErrors func(bool),
) *stateFactory {
	return &stateFactory{
		baseStateFactory:  baseStateFactory,
		scopeStateFactory: scopeStateFactory,
		provider:          provider,
		ignoreWatchErrors: ignoreWatchErrors,
	}
}

type stateFactory struct {
	baseStateFactory  composed.StateFactory
	scopeStateFactory commonscope.StateFactory
	provider          *cloudcontrolv1beta1.ProviderType
	ignoreWatchErrors func(bool)
}

func (f *stateFactory) NewState(ctx context.Context, req ctrl.Request) (*State, error) {
	scopeState, err := f.scopeStateFactory.NewState(
		ctx,
		req.NamespacedName,
		f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.CloudResources{}),
	)
	if err != nil {
		return nil, err
	}

	return &State{
		State:             scopeState,
		Provider:          f.provider,
		IgnoreWatchErrors: f.ignoreWatchErrors,
	}, nil
}

func (s *State) ObjAsCloudResources() *cloudresourcesv1beta1.CloudResources {
	return s.Obj().(*cloudresourcesv1beta1.CloudResources)
}
