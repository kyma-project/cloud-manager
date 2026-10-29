package awsredisinstance

import (
	"context"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestKcpRedisInstanceProvisioned(t *testing.T) {
	t.Run("false when KcpRedisInstance is nil", func(t *testing.T) {
		assert.False(t, kcpRedisInstanceProvisioned(context.Background(), &State{}))
	})

	t.Run("false when PrimaryEndpoint is not yet published", func(t *testing.T) {
		state := &State{KcpRedisInstance: &cloudcontrolv1beta1.RedisInstance{}}
		assert.False(t, kcpRedisInstanceProvisioned(context.Background(), state))
	})

	t.Run("true when PrimaryEndpoint is published", func(t *testing.T) {
		kcp := &cloudcontrolv1beta1.RedisInstance{}
		kcp.Status.PrimaryEndpoint = "host:6379"
		state := &State{KcpRedisInstance: kcp}
		assert.True(t, kcpRedisInstanceProvisioned(context.Background(), state))
	})

	// The endpoint persists across the Ready/Updating condition churn: during a
	// modify KCP clears the Ready condition but keeps Status.PrimaryEndpoint, so
	// the gate stays true and the secret keeps reconciling.
	t.Run("true with endpoint but no Ready condition (modify window)", func(t *testing.T) {
		kcp := &cloudcontrolv1beta1.RedisInstance{}
		kcp.Status.PrimaryEndpoint = "host:6379"
		state := &State{KcpRedisInstance: kcp}
		assert.True(t, kcpRedisInstanceProvisioned(context.Background(), state))
	})
}
