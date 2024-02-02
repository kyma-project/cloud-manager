package gcpnfsvolume

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type State struct {
	composed.State
	KymaRef        klog.ObjectRef
	KcpCluster     composed.StateCluster
	KcpNfsInstance *cloudcontrolv1beta1.NfsInstance
	KcpIpRange     *cloudcontrolv1beta1.IpRange
	SkrCluster     composed.StateCluster
	PV             *v1.PersistentVolume
}

type StateFactory interface {
	NewState(baseState composed.State) *State
}

func NewStateFactory(kymaRef klog.ObjectRef, kcpCluster composed.StateCluster, skrCluster composed.StateCluster) StateFactory {
	return &stateFactory{
		kymaRef:    kymaRef,
		kcpCluster: kcpCluster,
		skrCluster: skrCluster,
	}
}

type stateFactory struct {
	kymaRef    klog.ObjectRef
	kcpCluster composed.StateCluster
	skrCluster composed.StateCluster
}

func (f *stateFactory) NewState(baseState composed.State) *State {
	return &State{
		State:      baseState,
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
		SkrCluster: f.skrCluster,
	}
}

func (s *State) ObjAsGcpNfsVolume() *cloudresourcesv1beta1.GcpNfsVolume {
	return s.Obj().(*cloudresourcesv1beta1.GcpNfsVolume)
}

func (s *State) IsChanged() bool {
	return s.KcpNfsInstance != nil && s.KcpNfsInstance.Spec.Instance.Gcp.CapacityGb != s.ObjAsGcpNfsVolume().Spec.CapacityGb
}
