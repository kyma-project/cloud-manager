package awsredisinstance

import (
	"maps"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
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

	SkrIpRange       *cloudresourcesv1beta1.IpRange
	KcpRedisInstance *cloudcontrolv1beta1.RedisInstance

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
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AwsRedisInstance{}),
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
	}
}

func (s *State) ObjAsAwsRedisInstance() *cloudresourcesv1beta1.AwsRedisInstance {
	return s.Obj().(*cloudresourcesv1beta1.AwsRedisInstance)
}

func (s *State) GetSkrIpRange() *cloudresourcesv1beta1.IpRange {
	return s.SkrIpRange
}

func (s *State) SetSkrIpRange(skrIpRange *cloudresourcesv1beta1.IpRange) {
	s.SkrIpRange = skrIpRange
}

func (s *State) ObjAsObjWithIpRangeRef() defaultiprange.ObjWithIpRangeRef {
	return s.ObjAsAwsRedisInstance()
}

func (s *State) GetAuthSecretData() map[string][]byte {
	authSecretBaseData := getAuthSecretBaseData(s.KcpRedisInstance)
	redisInstance := s.ObjAsAwsRedisInstance()
	if redisInstance.Spec.AuthSecret == nil {
		return authSecretBaseData
	}

	parsedAuthSecretExtraData := parseAuthSecretExtraData(redisInstance.Spec.AuthSecret.ExtraData, authSecretBaseData)

	return util.MergeMaps(authSecretBaseData, parsedAuthSecretExtraData, false)
}

func (s *State) ShouldModifyKcp() bool {
	awsRedisInstance := s.ObjAsAwsRedisInstance()

	cacheNodeType, err := redisTierToCacheNodeTypeConvertor(awsRedisInstance.Spec.RedisTier)
	if err != nil {
		return true
	}

	areCacheNodeTypesDifferent := s.KcpRedisInstance.Spec.Instance.Aws.CacheNodeType != cacheNodeType
	isAutoMinorVersionUpgradeDifferent := s.KcpRedisInstance.Spec.Instance.Aws.AutoMinorVersionUpgrade != awsRedisInstance.Spec.AutoMinorVersionUpgrade
	isAuthEnabledDifferent := s.KcpRedisInstance.Spec.Instance.Aws.AuthEnabled != awsRedisInstance.Spec.AuthEnabled
	arePreferredMaintenanceWindowDifferent := ptr.Deref(s.KcpRedisInstance.Spec.Instance.Aws.PreferredMaintenanceWindow, "") != ptr.Deref(awsRedisInstance.Spec.PreferredMaintenanceWindow, "")

	return !maps.Equal(s.KcpRedisInstance.Spec.Instance.Aws.Parameters, awsRedisInstance.Spec.Parameters) ||
		areCacheNodeTypesDifferent ||
		isAutoMinorVersionUpgradeDifferent ||
		isAuthEnabledDifferent ||
		arePreferredMaintenanceWindowDifferent
}
