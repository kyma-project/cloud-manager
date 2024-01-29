package crgcpnfsvolume

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/components/kcp/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/components/kcp/pkg/composed"
	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef        klog.ObjectRef
	KcpCluster     composed.StateCluster
	KcpNfsInstance *cloudcontrolv1beta1.NfsInstance
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

func (s *State) ObjAsGcpNfsVolume() *cloudresourcesv1beta1.GcpNfsVolume {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolume)
}
