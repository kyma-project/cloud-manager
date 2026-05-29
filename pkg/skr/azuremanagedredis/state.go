package azuremanagedredis

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultiprange"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
)

type State struct {
	composed.State
	KymaRef    klog.ObjectRef
	KcpCluster composed.StateCluster

	KcpAzureManagedRedis *cloudcontrolv1beta1.AzureManagedRedis
	SkrIpRange           *cloudresourcesv1beta1.IpRange
	AuthSecret           *corev1.Secret
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
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AzureManagedRedis{}),
		KymaRef:    kymaRef,
		KcpCluster: f.kcpCluster,
	}, nil
}

func (s *State) ObjAsAzureManagedRedis() *cloudresourcesv1beta1.AzureManagedRedis {
	return s.Obj().(*cloudresourcesv1beta1.AzureManagedRedis)
}

func (s *State) ObjAsObjWithIpRangeRef() defaultiprange.ObjWithIpRangeRef {
	return s.ObjAsAzureManagedRedis()
}

func (s *State) GetSkrIpRange() *cloudresourcesv1beta1.IpRange {
	return s.SkrIpRange
}

func (s *State) SetSkrIpRange(skrIpRange *cloudresourcesv1beta1.IpRange) {
	s.SkrIpRange = skrIpRange
}

func (s *State) GetAuthSecretData() map[string][]byte {
	authSecretBaseData := getAuthSecretBaseData(s.KcpAzureManagedRedis)
	amr := s.ObjAsAzureManagedRedis()
	if amr.Spec.AuthSecret == nil {
		return authSecretBaseData
	}

	parsedAuthSecretExtraData := parseAuthSecretExtraData(amr.Spec.AuthSecret.ExtraData, authSecretBaseData)

	return util.MergeMaps(authSecretBaseData, parsedAuthSecretExtraData, false)
}
