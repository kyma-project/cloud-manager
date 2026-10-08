package alicloudrediscluster

import (
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestRedisTierToInstanceClass(t *testing.T) {
	tests := []struct {
		tier                  cloudresourcesv1beta1.AlicloudRedisClusterTier
		expectedInstanceClass string
		expectError           bool
	}{
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC3, "redis.shard.small.ce", false},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC4, "redis.shard.mid.ce", false},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC5, "redis.shard.large.ce", false},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC6, "redis.shard.2xlarge.ce", false},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC7, "redis.shard.4xlarge.ce", false},
		{"unknown", "", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			instanceClass, err := redisTierToInstanceClass(tt.tier)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedInstanceClass, instanceClass)
			}
		})
	}
}

func TestClusterTiersAreOrdered(t *testing.T) {
	expected := []struct {
		tier  cloudresourcesv1beta1.AlicloudRedisClusterTier
		class string
	}{
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC3, "redis.shard.small.ce"},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC4, "redis.shard.mid.ce"},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC5, "redis.shard.large.ce"},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC6, "redis.shard.2xlarge.ce"},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC7, "redis.shard.4xlarge.ce"},
	}

	for _, tc := range expected {
		got, err := redisTierToInstanceClass(tc.tier)
		assert.NoError(t, err)
		assert.Equal(t, tc.class, got, "tier %s", tc.tier)
	}
}
