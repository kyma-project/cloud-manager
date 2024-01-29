package criprange

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster
	KcpIpRange *cloudcontrolv1beta1.IpRange
}

type StateFactory interface {
	NewState(baseState composed.State) *State
}

func NewStateFactory(kymaRef klog.ObjectRef, kcpCluster composed.StateCluster) StateFactory {
	return &stateFactory{
		kymaRef:    kymaRef,
		kcpCluster: kcpCluster,
	}
}

type stateFactory struct {
	kymaRef    klog.ObjectRef
	kcpCluster composed.StateCluster
}

func (f *stateFactory) NewState(baseState composed.State) *State {
	return &State{
		State:      baseState,
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
	}
}

func (s *State) ObjAsIpRange() *cloudresourcesv1beta1.IpRange {
	return s.Obj().(*cloudresourcesv1beta1.IpRange)
}
