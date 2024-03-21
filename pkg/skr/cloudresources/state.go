package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	composed.State

	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	Provider   *cloudcontrolv1beta1.ProviderType
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	kymaRef klog.ObjectRef,
	kcpCluster composed.StateCluster,
	provider *cloudcontrolv1beta1.ProviderType,
) *stateFactory {
	return &stateFactory{
		baseStateFactory: baseStateFactory,
		kymaRef:          kymaRef,
		kcpCluster:       kcpCluster,
		provider:         provider,
	}
}

type stateFactory struct {
	baseStateFactory composed.StateFactory
	kymaRef          klog.ObjectRef
	kcpCluster       composed.StateCluster
	provider         *cloudcontrolv1beta1.ProviderType
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	return &State{
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.CloudResources{}),
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
		Provider:   f.provider,
	}
}

func (s *State) ObjAsCloudResources() *cloudresourcesv1beta1.CloudResources {
	return s.Obj().(*cloudresourcesv1beta1.CloudResources)
}
