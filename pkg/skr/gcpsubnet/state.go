package gcpsubnet

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

	KcpGcpSubnet *cloudcontrolv1beta1.GcpSubnet
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	kymaRef klog.ObjectRef,
	kcpCluster composed.StateCluster,
) *stateFactory {
	return &stateFactory{
		baseStateFactory: baseStateFactory,
		kymaRef:          kymaRef,
		kcpCluster:       kcpCluster,
	}
}

type stateFactory struct {
	baseStateFactory composed.StateFactory
	kymaRef          klog.ObjectRef
	kcpCluster       composed.StateCluster
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	return &State{
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.GcpSubnet{}),
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
	}
}

func (s *State) ObjAsGcpSubnet() *cloudresourcesv1beta1.GcpSubnet {
	return s.Obj().(*cloudresourcesv1beta1.GcpSubnet)
}
