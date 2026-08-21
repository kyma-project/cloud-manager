package alicloudrediscluster

import (
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestRedisTierToInstanceClass(t *testing.T) {
	tests := []struct {
		tier                  cloudresourcesv1beta1.AlicloudRedisClusterTier
		shardCount            int32
		expectedInstanceClass string
		expectError           bool
	}{
		// shardCount=1 → proxyCount clamped to 4
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC3, 1, "redis.logic.sharding.4g.1db.0rodb.4proxy.default", false},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC4, 1, "redis.logic.sharding.8g.1db.0rodb.4proxy.default", false},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC5, 1, "redis.logic.sharding.16g.1db.0rodb.4proxy.default", false},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC6, 1, "redis.logic.sharding.32g.1db.0rodb.4proxy.default", false},
		// shardCount=8 → proxyCount=8
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC3, 8, "redis.logic.sharding.4g.8db.0rodb.8proxy.default", false},
		// unknown tier
		{"unknown", 1, "", true},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			instanceClass, err := redisTierToInstanceClass(tt.tier, tt.shardCount)
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
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC3, "redis.logic.sharding.4g.1db.0rodb.4proxy.default"},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC4, "redis.logic.sharding.8g.1db.0rodb.4proxy.default"},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC5, "redis.logic.sharding.16g.1db.0rodb.4proxy.default"},
		{cloudresourcesv1beta1.AlicloudRedisClusterTierC6, "redis.logic.sharding.32g.1db.0rodb.4proxy.default"},
	}

	for _, tc := range expected {
		got, err := redisTierToInstanceClass(tc.tier, 1)
		assert.NoError(t, err)
		assert.Equal(t, tc.class, got, "tier %s", tc.tier)
	}
}
