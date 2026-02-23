package azurerediscluster

import (
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

	KcpRedisCluster *cloudcontrolv1beta1.RedisCluster
	SkrIpRange      *cloudresourcesv1beta1.IpRange
	AuthSecret      *corev1.Secret
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
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.AzureRedisCluster{}),
		KymaRef:    f.kymaRef,
		KcpCluster: f.kcpCluster,
	}
}

func (s *State) ObjAsAzureRedisCluster() *cloudresourcesv1beta1.AzureRedisCluster {
	return s.Obj().(*cloudresourcesv1beta1.AzureRedisCluster)
}

func (s *State) ObjAsObjWithIpRangeRef() defaultiprange.ObjWithIpRangeRef {
	return s.ObjAsAzureRedisCluster()
}

func (s *State) GetSkrIpRange() *cloudresourcesv1beta1.IpRange {
	return s.SkrIpRange
}

func (s *State) SetSkrIpRange(skrIpRange *cloudresourcesv1beta1.IpRange) {
	s.SkrIpRange = skrIpRange
}

func (s *State) GetAuthSecretData() map[string][]byte {
	authSecretBaseData := getAuthSecretBaseData(s.KcpRedisCluster)
	redisCluster := s.ObjAsAzureRedisCluster()

	effectiveAuthSecret := getEffectiveAuthSecret(redisCluster)
	if effectiveAuthSecret == nil {
		return authSecretBaseData
	}

	parsedAuthSecretExtraData := parseAuthSecretExtraData(effectiveAuthSecret.ExtraData, authSecretBaseData)

	return util.MergeMaps(authSecretBaseData, parsedAuthSecretExtraData, false)
}
