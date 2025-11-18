package awsrediscluster

import (
	"maps"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	awsconfig "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/config"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultiprange"
	"github.com/kyma-project/cloud-manager/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"

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
	kymaRef klog.ObjectRef,
	kcpCluster composed.StateCluster,
) *stateFactory {
	return &stateFactory{
		baseStateFactory: baseStateFactory,
		kymaRef:          kymaRef,
		kcpCluster:       kcpCluster,
	}
}

type stateFactory struct {
	baseStateFactory composed.StateFactory
	kymaRef          klog.ObjectRef
	kcpCluster       composed.StateCluster
}

func (f *stateFactory) NewState(req ctrl.Request) *State {
	return &State{
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsRedisCluster{}),
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
	}
}

func (s *State) ObjAsAwsRedisCluster() *cloudresourcesv1beta1.AwsRedisCluster {
	return s.Obj().(*cloudresourcesv1beta1.AwsRedisCluster)
}

func (s *State) GetSkrIpRange() *cloudresourcesv1beta1.IpRange {
	return s.SkrIpRange
}

func (s *State) SetSkrIpRange(skrIpRange *cloudresourcesv1beta1.IpRange) {
	s.SkrIpRange = skrIpRange
}

func (s *State) ObjAsObjWithIpRangeRef() defaultiprange.ObjWithIpRangeRef {
	return s.ObjAsAwsRedisCluster()
}

func (s *State) GetAuthSecretData() map[string][]byte {
	authSecretBaseData := getAuthSecretBaseData(s.KcpRedisCluster)
	redisCluster := s.ObjAsAwsRedisCluster()
	if redisCluster.Spec.AuthSecret == nil {
		return authSecretBaseData
	}

	parsedAuthSecretExtraData := parseAuthSecretExtraData(redisCluster.Spec.AuthSecret.ExtraData, authSecretBaseData)

	return util.MergeMaps(authSecretBaseData, parsedAuthSecretExtraData, false)
}

func (s *State) ShouldModifyKcp() bool {
	awsRedisCluster := s.ObjAsAwsRedisCluster()

	cacheNodeType, err := redisTierToCacheNodeTypeConvertor(awsRedisCluster.Spec.RedisTier, awsconfig.AwsConfig.RedisClusterTierMachineTypes)
	if err != nil {
		return true
	}

	areCacheNodeTypesDifferent := s.KcpRedisCluster.Spec.Instance.Aws.CacheNodeType != cacheNodeType
	isAutoMinorVersionUpgradeDifferent := s.KcpRedisCluster.Spec.Instance.Aws.AutoMinorVersionUpgrade != awsRedisCluster.Spec.AutoMinorVersionUpgrade
	isAuthEnabledDifferent := s.KcpRedisCluster.Spec.Instance.Aws.AuthEnabled != awsRedisCluster.Spec.AuthEnabled
	arePreferredMaintenanceWindowDifferent := ptr.Deref(s.KcpRedisCluster.Spec.Instance.Aws.PreferredMaintenanceWindow, "") != ptr.Deref(awsRedisCluster.Spec.PreferredMaintenanceWindow, "")
	isEngineVersionDifferent := s.KcpRedisCluster.Spec.Instance.Aws.EngineVersion != awsRedisCluster.Spec.EngineVersion
	isShardCountDifferent := s.KcpRedisCluster.Spec.Instance.Aws.ShardCount != awsRedisCluster.Spec.ShardCount
	isReplicaCountDifferent := s.KcpRedisCluster.Spec.Instance.Aws.ReplicasPerShard != awsRedisCluster.Spec.ReplicasPerShard

	return !maps.Equal(s.KcpRedisCluster.Spec.Instance.Aws.Parameters, awsRedisCluster.Spec.Parameters) ||
		areCacheNodeTypesDifferent ||
		isAutoMinorVersionUpgradeDifferent ||
		isAuthEnabledDifferent ||
		arePreferredMaintenanceWindowDifferent ||
		isEngineVersionDifferent ||
		isShardCountDifferent ||
		isReplicaCountDifferent
}
