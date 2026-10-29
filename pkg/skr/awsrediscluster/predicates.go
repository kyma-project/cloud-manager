package awsrediscluster

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

// kcpRedisClusterReady reports whether the KCP RedisCluster has published its Ready condition (and thus its connection data).
func kcpRedisClusterReady(_ context.Context, st composed.State) bool {
	state := st.(*State)
	return state.KcpRedisCluster != nil &&
		meta.FindStatusCondition(state.KcpRedisCluster.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
}
