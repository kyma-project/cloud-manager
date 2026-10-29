package awsredisinstance

import (
	"context"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestKcpRedisInstanceReady(t *testing.T) {
	t.Run("false when KcpRedisInstance is nil", func(t *testing.T) {
		state := &State{}
		assert.False(t, kcpRedisInstanceReady(context.Background(), state))
	})

	t.Run("false when KcpRedisInstance has no Ready condition", func(t *testing.T) {
		state := &State{KcpRedisInstance: &cloudcontrolv1beta1.RedisInstance{}}
		assert.False(t, kcpRedisInstanceReady(context.Background(), state))
	})

	t.Run("true when Ready condition has ConditionTrue", func(t *testing.T) {
		kcp := &cloudcontrolv1beta1.RedisInstance{}
		kcp.Status.Conditions = []metav1.Condition{{
			Type:   cloudcontrolv1beta1.ConditionTypeReady,
			Status: metav1.ConditionTrue,
		}}
		state := &State{KcpRedisInstance: kcp}
		assert.True(t, kcpRedisInstanceReady(context.Background(), state))
	})

	// Presence-only by design, matching updateStatus's own Ready check: KCP only
	// ever publishes Ready with Status=True (SetExclusiveConditions), so presence
	// is equivalent to True in practice. This pins that intent.
	t.Run("true when Ready condition is present even with ConditionFalse (presence-only)", func(t *testing.T) {
		kcp := &cloudcontrolv1beta1.RedisInstance{}
		kcp.Status.Conditions = []metav1.Condition{{
			Type:   cloudcontrolv1beta1.ConditionTypeReady,
			Status: metav1.ConditionFalse,
		}}
		state := &State{KcpRedisInstance: kcp}
		assert.True(t, kcpRedisInstanceReady(context.Background(), state))
	})
}
