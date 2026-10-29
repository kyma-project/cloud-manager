package awsredisinstance

import (
	"context"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
)

// kcpRedisInstanceReady reports whether the KCP RedisInstance has published its Ready condition (and thus its connection data).
func kcpRedisInstanceReady(_ context.Context, st composed.State) bool {
	state := st.(*State)
	return state.KcpRedisInstance != nil &&
		meta.FindStatusCondition(state.KcpRedisInstance.Status.Conditions, cloudcontrolv1beta1.ConditionTypeReady) != nil
}
