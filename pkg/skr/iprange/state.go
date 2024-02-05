package iprange

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	skrruntime "github.com/kyma-project/cloud-manager/pkg/skr/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	Reloader   skrruntime.Reloader

	KcpIpRange *cloudcontrolv1beta1.IpRange
}

func newStateFactory(baseStateFactory composed.StateFactory, kymaRef klog.ObjectRef, kcpCluster composed.StateCluster, reloader skrruntime.Reloader) *stateFactory {
	return &stateFactory{
		baseStateFactory: baseStateFactory,
		kymaRef:          kymaRef,
		kcpCluster:       kcpCluster,
		reloader:         reloader,
	}
}

type stateFactory struct {
	baseStateFactory composed.StateFactory
	kymaRef          klog.ObjectRef
	kcpCluster       composed.StateCluster
	reloader         skrruntime.Reloader
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	return &State{
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.IpRange{}),
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
		Reloader:   f.reloader,
	}
}

func (s *State) ObjAsIpRange() *cloudresourcesv1beta1.IpRange {
	return s.Obj().(*cloudresourcesv1beta1.IpRange)
}
