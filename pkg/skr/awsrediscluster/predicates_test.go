package awsrediscluster

import (
	"context"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKcpRedisClusterReady(t *testing.T) {
	t.Run("false when KcpRedisCluster is nil", func(t *testing.T) {
		state := &State{}
		assert.False(t, kcpRedisClusterReady(context.Background(), state))
	})

	t.Run("false when KcpRedisCluster has no Ready condition", func(t *testing.T) {
		state := &State{KcpRedisCluster: &cloudcontrolv1beta1.RedisCluster{}}
		assert.False(t, kcpRedisClusterReady(context.Background(), state))
	})

	t.Run("true when Ready condition has ConditionTrue", func(t *testing.T) {
		kcp := &cloudcontrolv1beta1.RedisCluster{}
		kcp.Status.Conditions = []metav1.Condition{{
			Type:   cloudcontrolv1beta1.ConditionTypeReady,
			Status: metav1.ConditionTrue,
		}}
		state := &State{KcpRedisCluster: kcp}
		assert.True(t, kcpRedisClusterReady(context.Background(), state))
	})

	// Presence-only by design, matching updateStatus's own Ready check: KCP only
	// ever publishes Ready with Status=True (SetExclusiveConditions), so presence
	// is equivalent to True in practice. This pins that intent.
	t.Run("true when Ready condition is present even with ConditionFalse (presence-only)", func(t *testing.T) {
		kcp := &cloudcontrolv1beta1.RedisCluster{}
		kcp.Status.Conditions = []metav1.Condition{{
			Type:   cloudcontrolv1beta1.ConditionTypeReady,
			Status: metav1.ConditionFalse,
		}}
		state := &State{KcpRedisCluster: kcp}
		assert.True(t, kcpRedisClusterReady(context.Background(), state))
	})
}
