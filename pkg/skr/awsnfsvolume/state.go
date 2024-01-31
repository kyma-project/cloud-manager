package awsnfsvolume

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

	SkrIpRange     *cloudresourcesv1beta1.IpRange
	KcpNfsInstance *cloudcontrolv1beta1.NfsInstance
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	kymaRef klog.ObjectRef,
	kcpCluster composed.StateCluster,
) *stateFactory {
	return &stateFactory{
		baseStateFactory: nil,
		kymaRef:          klog.ObjectRef{},
		kcpCluster:       nil,
	}
}

type stateFactory struct {
	baseStateFactory composed.StateFactory
	kymaRef          klog.ObjectRef
	kcpCluster       composed.StateCluster
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	return newState(
		f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsNfsVolume{}),
		f.kymaRef,
		f.kcpCluster,
	)
}

func newState(baseState composed.State, kymaRef klog.ObjectRef, kcpCluster composed.StateCluster) *State {
	return &State{
		State:      baseState,
		KymaRef:    kymaRef,
		KcpCluster: kcpCluster,
	}
}

func (s *State) ObjAsAwsNfsVolume() *cloudresourcesv1beta1.AwsNfsVolume {
	return s.Obj().(*cloudresourcesv1beta1.AwsNfsVolume)
}
