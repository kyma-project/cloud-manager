package awsvpcpeering

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster

	KcpVpcPeering *cloudcontrolv1beta1.VpcPeering
	RemoteNetwork *cloudcontrolv1beta1.Network
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
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsVpcPeering{}),
		KymaRef:    kymaRef,
		KcpCluster: f.kcpCluster,
	}, nil
}

func (s *State) ObjAsAwsVpcPeering() *cloudresourcesv1beta1.AwsVpcPeering {
	return s.Obj().(*cloudresourcesv1beta1.AwsVpcPeering)
}
