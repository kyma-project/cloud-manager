package gcpredisinstance

import (
	"maps"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultiprange"
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
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.GcpRedisInstance{}),
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
	}
}

func (s *State) ObjAsGcpRedisInstance() *cloudresourcesv1beta1.GcpRedisInstance {
	return s.Obj().(*cloudresourcesv1beta1.GcpRedisInstance)
}

func (s *State) GetSkrIpRange() *cloudresourcesv1beta1.IpRange {
	return s.SkrIpRange
}

func (s *State) SetSkrIpRange(skrIpRange *cloudresourcesv1beta1.IpRange) {
	s.SkrIpRange = skrIpRange
}

func (s *State) ObjAsObjWithIpRangeRef() defaultiprange.ObjWithIpRangeRef {
	return s.ObjAsGcpRedisInstance()
}

func (s *State) ShouldModifyKcp() bool {
	gcpRedisInstance := s.ObjAsGcpRedisInstance()

	_, memorySizeGb, err := redisTierToTierAndMemorySizeConverter(gcpRedisInstance.Spec.RedisTier)
	if err != nil {
		return true
	}

	areMemorySizesGbDifferent := s.KcpRedisInstance.Spec.Instance.Gcp.MemorySizeGb != memorySizeGb
	areAuthEnablesDifferent := s.KcpRedisInstance.Spec.Instance.Gcp.AuthEnabled != gcpRedisInstance.Spec.AuthEnabled
	areRedisVersionsDifferent := s.KcpRedisInstance.Spec.Instance.Gcp.RedisVersion != gcpRedisInstance.Spec.RedisVersion

	return !maps.Equal(s.KcpRedisInstance.Spec.Instance.Gcp.RedisConfigs, gcpRedisInstance.Spec.RedisConfigs) ||
		areMemorySizesGbDifferent ||
		areMaintenancePoliciesDifferent(gcpRedisInstance.Spec.MaintenancePolicy, s.KcpRedisInstance.Spec.Instance.Gcp.MaintenancePolicy) ||
		areAuthEnablesDifferent ||
		areRedisVersionsDifferent
}

func (s *State) GetAuthSecretData() map[string][]byte {
	authSecretBaseData := getAuthSecretBaseData(s.KcpRedisInstance)
	redisInstance := s.ObjAsGcpRedisInstance()
	if redisInstance.Spec.AuthSecret == nil {
		return authSecretBaseData
	}

	parsedAuthSecretExtraData := parseAuthSecretExtraData(redisInstance.Spec.AuthSecret.ExtraData, authSecretBaseData)

	return util.MergeMaps(authSecretBaseData, parsedAuthSecretExtraData, false)
}

func areMaintenancePoliciesDifferent(skrPolicy *cloudresourcesv1beta1.MaintenancePolicy, kcpPolicy *cloudcontrolv1beta1.MaintenancePolicyGcp) bool {
	if skrPolicy == nil && kcpPolicy == nil {
		return false
	}

	if (skrPolicy == nil && kcpPolicy != nil) || (skrPolicy != nil && kcpPolicy == nil) {
		return true
	}

	if skrPolicy.DayOfWeek == nil && kcpPolicy.DayOfWeek == nil {
		return false
	}

	if (skrPolicy.DayOfWeek == nil && kcpPolicy.DayOfWeek != nil) || (skrPolicy.DayOfWeek != nil && kcpPolicy.DayOfWeek == nil) {
		return true
	}

	if skrPolicy.DayOfWeek.Day != kcpPolicy.DayOfWeek.Day {
		return true
	}

	if skrPolicy.DayOfWeek.StartTime.Hours != kcpPolicy.DayOfWeek.StartTime.Hours {
		return true
	}

	if skrPolicy.DayOfWeek.StartTime.Minutes != kcpPolicy.DayOfWeek.StartTime.Minutes {
		return true
	}

	return false
}
