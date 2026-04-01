package awsnfsvolume

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultiprange"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster

	SkrIpRange     *cloudresourcesv1beta1.IpRange
	KcpNfsInstance *cloudcontrolv1beta1.NfsInstance
	Volume         *corev1.PersistentVolume
	PVC            *corev1.PersistentVolumeClaim
}

func newStateFactory(
	baseStateFactory composed.StateFactory,
	scopeProvider scopeprovider.ScopeProvider,
	kcpCluster composed.StateCluster,
) *stateFactory {
	return &stateFactory{
		baseStateFactory: baseStateFactory,
		scopeProvider:    scopeProvider,
		kcpCluster:       kcpCluster,
	}
}

type stateFactory struct {
	baseStateFactory composed.StateFactory
	scopeProvider    scopeprovider.ScopeProvider
	kcpCluster       composed.StateCluster
}

func (f *stateFactory) NewState(ctx context.Context, req ctrl.Request) (*State, error) {
	kymaRef, err := f.scopeProvider.GetScope(ctx, req.NamespacedName)
	if err != nil {
		return nil, err
	}
	return &State{
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsNfsVolume{}),
		KymaRef:    kymaRef,
		KcpCluster: f.kcpCluster,
	}, nil
}

func (s *State) ObjAsAwsNfsVolume() *cloudresourcesv1beta1.AwsNfsVolume {
	return s.Obj().(*cloudresourcesv1beta1.AwsNfsVolume)
}

func (s *State) GetSkrIpRange() *cloudresourcesv1beta1.IpRange {
	return s.SkrIpRange
}

func (s *State) SetSkrIpRange(skrIpRange *cloudresourcesv1beta1.IpRange) {
	s.SkrIpRange = skrIpRange
}

func (s *State) ObjAsObjWithIpRangeRef() defaultiprange.ObjWithIpRangeRef {
	return s.ObjAsAwsNfsVolume()
}
