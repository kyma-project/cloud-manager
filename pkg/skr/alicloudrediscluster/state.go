package alicloudrediscluster

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

	SkrIpRange      *cloudresourcesv1beta1.IpRange
	KcpRedisCluster *cloudcontrolv1beta1.RedisCluster

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
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AlicloudRedisCluster{}),
		KymaRef:    kymaRef,
		KcpCluster: f.kcpCluster,
	}, nil
}

func (s *State) ObjAsAlicloudRedisCluster() *cloudresourcesv1beta1.AlicloudRedisCluster {
	return s.Obj().(*cloudresourcesv1beta1.AlicloudRedisCluster)
}

func (s *State) GetSkrIpRange() *cloudresourcesv1beta1.IpRange {
	return s.SkrIpRange
}

func (s *State) SetSkrIpRange(skrIpRange *cloudresourcesv1beta1.IpRange) {
	s.SkrIpRange = skrIpRange
}

func (s *State) ObjAsObjWithIpRangeRef() defaultiprange.ObjWithIpRangeRef {
	return s.ObjAsAlicloudRedisCluster()
}

func (s *State) GetAuthSecretData() map[string][]byte {
	authSecretBaseData := getAuthSecretBaseData(s.KcpRedisCluster)
	redisCluster := s.ObjAsAlicloudRedisCluster()
	if redisCluster.Spec.AuthSecret == nil {
		return authSecretBaseData
	}

	parsedAuthSecretExtraData := parseAuthSecretExtraData(redisCluster.Spec.AuthSecret.ExtraData, authSecretBaseData)

	return util.MergeMaps(authSecretBaseData, parsedAuthSecretExtraData, false)
}

func (s *State) ShouldModifyKcp() bool {
	alicloudRedisCluster := s.ObjAsAlicloudRedisCluster()

	instanceClass, err := redisTierToInstanceClass(alicloudRedisCluster.Spec.RedisTier, alicloudRedisCluster.Spec.ShardCount)
	if err != nil {
		// Unknown tier — trigger a KCP modify so the reconciler surfaces the error
		// rather than silently skipping the drift check.
		return true
	}

	if s.KcpRedisCluster.Spec.Instance.Alicloud == nil {
		return true
	}

	// Compare tier keys rather than full class strings: for proxy classes the
	// class name encodes shardCount, so a shard-count-only change must not be
	// treated as a tier (instanceClass) change here — isShardCountDifferent
	// already captures that.
	isInstanceClassDifferent := proxyClassTierKey(s.KcpRedisCluster.Spec.Instance.Alicloud.InstanceClass) != proxyClassTierKey(instanceClass)
	isShardCountDifferent := s.KcpRedisCluster.Spec.Instance.Alicloud.ShardCount != alicloudRedisCluster.Spec.ShardCount
	isReplicasPerShardDifferent := s.KcpRedisCluster.Spec.Instance.Alicloud.ReplicasPerShard != alicloudRedisCluster.Spec.ReplicasPerShard

	return !maps.Equal(s.KcpRedisCluster.Spec.Instance.Alicloud.Parameters, alicloudRedisCluster.Spec.Parameters) ||
		isInstanceClassDifferent ||
		isShardCountDifferent ||
		isReplicasPerShardDifferent
}
