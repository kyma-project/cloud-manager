package gcprediscluster

import (
	"context"
	"maps"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/skr/common/defaultgcpsubnet"
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

	SkrSubnet          *cloudresourcesv1beta1.GcpSubnet
	KcpGcpRedisCluster *cloudcontrolv1beta1.GcpRedisCluster

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
		State:      f.baseStateFactory.NewState(req.NamespacedName, &cloudresourcesv1beta1.GcpRedisCluster{}),
		KymaRef:    kymaRef,
		KcpCluster: f.kcpCluster,
	}, nil
}

func (s *State) ObjAsGcpRedisCluster() *cloudresourcesv1beta1.GcpRedisCluster {
	return s.Obj().(*cloudresourcesv1beta1.GcpRedisCluster)
}

func (s *State) GetSkrGcpSubnet() *cloudresourcesv1beta1.GcpSubnet {
	return s.SkrSubnet
}

func (s *State) SetSkrGcpSubnet(skrSubnet *cloudresourcesv1beta1.GcpSubnet) {
	s.SkrSubnet = skrSubnet
}

func (s *State) ObjAsObjWithGcpSubnetRef() defaultgcpsubnet.ObjWithGcpSubnetRef {
	return s.ObjAsGcpRedisCluster()
}

func (s *State) ShouldModifyKcp() bool {
	gcpRedisCluster := s.ObjAsGcpRedisCluster()

	nodeType, err := redisTierToNodeTypeConverter(gcpRedisCluster.Spec.RedisTier)
	if err != nil {
		return true
	}

	areMemorySizesGbDifferent := s.KcpGcpRedisCluster.Spec.NodeType != nodeType
	isReplicaCountDifferent := s.KcpGcpRedisCluster.Spec.ReplicasPerShard != gcpRedisCluster.Spec.ReplicasPerShard
	isShardCountDifferent := s.KcpGcpRedisCluster.Spec.ShardCount != gcpRedisCluster.Spec.ShardCount

	return !maps.Equal(s.KcpGcpRedisCluster.Spec.RedisConfigs, gcpRedisCluster.Spec.RedisConfigs) ||
		areMemorySizesGbDifferent ||
		isShardCountDifferent ||
		isReplicaCountDifferent
}

func (s *State) GetAuthSecretData() map[string][]byte {
	authSecretBaseData := getAuthSecretBaseData(s.KcpGcpRedisCluster)
	redisCluster := s.ObjAsGcpRedisCluster()
	if redisCluster.Spec.AuthSecret == nil {
		return authSecretBaseData
	}

	parsedAuthSecretExtraData := parseAuthSecretExtraData(redisCluster.Spec.AuthSecret.ExtraData, authSecretBaseData)

	return util.MergeMaps(authSecretBaseData, parsedAuthSecretExtraData, false)
}
