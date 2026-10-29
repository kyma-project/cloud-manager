package awsrediscluster

import (
	"context"
	"testing"

	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestKcpRedisClusterProvisioned(t *testing.T) {
	t.Run("false when KcpRedisCluster is nil", func(t *testing.T) {
		assert.False(t, kcpRedisClusterProvisioned(context.Background(), &State{}))
	})

	t.Run("false when DiscoveryEndpoint is not yet published", func(t *testing.T) {
		state := &State{KcpRedisCluster: &cloudcontrolv1beta1.RedisCluster{}}
		assert.False(t, kcpRedisClusterProvisioned(context.Background(), state))
	})

	t.Run("true when DiscoveryEndpoint is published", func(t *testing.T) {
		kcp := &cloudcontrolv1beta1.RedisCluster{}
		kcp.Status.DiscoveryEndpoint = "host:6379"
		state := &State{KcpRedisCluster: kcp}
		assert.True(t, kcpRedisClusterProvisioned(context.Background(), state))
	})

	// The endpoint persists across the Ready/Updating condition churn: during a
	// modify KCP clears the Ready condition but keeps Status.DiscoveryEndpoint, so
	// the gate stays true and the secret keeps reconciling.
	t.Run("true with endpoint but no Ready condition (modify window)", func(t *testing.T) {
		kcp := &cloudcontrolv1beta1.RedisCluster{}
		kcp.Status.DiscoveryEndpoint = "host:6379"
		state := &State{KcpRedisCluster: kcp}
		assert.True(t, kcpRedisClusterProvisioned(context.Background(), state))
	})
}
