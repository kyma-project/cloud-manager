package alicloudredisinstance

import (
	"context"
	"maps"

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

	SkrIpRange       *cloudresourcesv1beta1.IpRange
	KcpRedisInstance *cloudcontrolv1beta1.RedisInstance

	AuthSecret *corev1.Secret
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
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AlicloudRedisInstance{}),
		KymaRef:    kymaRef,
		KcpCluster: f.kcpCluster,
	}, nil
}

func (s *State) ObjAsAlicloudRedisInstance() *cloudresourcesv1beta1.AlicloudRedisInstance {
	return s.Obj().(*cloudresourcesv1beta1.AlicloudRedisInstance)
}

func (s *State) GetSkrIpRange() *cloudresourcesv1beta1.IpRange {
	return s.SkrIpRange
}

func (s *State) SetSkrIpRange(skrIpRange *cloudresourcesv1beta1.IpRange) {
	s.SkrIpRange = skrIpRange
}

func (s *State) ObjAsObjWithIpRangeRef() defaultiprange.ObjWithIpRangeRef {
	return s.ObjAsAlicloudRedisInstance()
}

func (s *State) GetAuthSecretData() map[string][]byte {
	authSecretBaseData := getAuthSecretBaseData(s.KcpRedisInstance)
	redisInstance := s.ObjAsAlicloudRedisInstance()
	if redisInstance.Spec.AuthSecret == nil {
		return authSecretBaseData
	}

	parsedAuthSecretExtraData := parseAuthSecretExtraData(redisInstance.Spec.AuthSecret.ExtraData, authSecretBaseData)

	return util.MergeMaps(authSecretBaseData, parsedAuthSecretExtraData, false)
}

func (s *State) ShouldModifyKcp() bool {
	alicloudRedisInstance := s.ObjAsAlicloudRedisInstance()

	instanceClass, readOnlyCount, err := redisTierToInstanceClassAndReadOnlyCount(alicloudRedisInstance.Spec.RedisTier)
	if err != nil {
		// Unknown tier — trigger a KCP modify so the reconciler surfaces the error
		// rather than silently skipping the drift check.
		return true
	}

	if s.KcpRedisInstance.Spec.Instance.Alicloud == nil {
		return true
	}

	isInstanceClassDifferent := s.KcpRedisInstance.Spec.Instance.Alicloud.InstanceClass != instanceClass
	isReadOnlyCountDifferent := s.KcpRedisInstance.Spec.Instance.Alicloud.ReadOnlyCount != readOnlyCount

	return !maps.Equal(s.KcpRedisInstance.Spec.Instance.Alicloud.Parameters, alicloudRedisInstance.Spec.Parameters) ||
		isInstanceClassDifferent ||
		isReadOnlyCountDifferent
}
