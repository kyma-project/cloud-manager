package alicloudredisinstance

import (
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestRedisTierToInstanceClassAndReadOnlyCount(t *testing.T) {
	tests := []struct {
		tier                  cloudresourcesv1beta1.AlicloudRedisTier
		expectedInstanceClass string
		expectedReadOnlyCount int32
		expectError           bool
	}{
		{cloudresourcesv1beta1.AlicloudRedisTierS1, "redis.master.small.cloud", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS2, "redis.master.mid.cloud", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS3, "redis.master.stand.cloud", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS4, "redis.master.large.cloud", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS5, "redis.master.2xlarge.cloud", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP1, "redis.master.stand.cloud", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP2, "redis.master.large.cloud", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP3, "redis.master.2xlarge.cloud", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP4, "redis.master.4xlarge.cloud", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP5, "redis.master.8xlarge.cloud", 1, false},
		{"unknown", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.tier), func(t *testing.T) {
			instanceClass, readOnlyCount, err := redisTierToInstanceClassAndReadOnlyCount(tt.tier)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedInstanceClass, instanceClass)
				assert.Equal(t, tt.expectedReadOnlyCount, readOnlyCount)
			}
		})
	}
}

func TestPTiersUseReadOnlyReplica(t *testing.T) {
	sTiers := []cloudresourcesv1beta1.AlicloudRedisTier{
		cloudresourcesv1beta1.AlicloudRedisTierS1,
		cloudresourcesv1beta1.AlicloudRedisTierS2,
		cloudresourcesv1beta1.AlicloudRedisTierS3,
		cloudresourcesv1beta1.AlicloudRedisTierS4,
		cloudresourcesv1beta1.AlicloudRedisTierS5,
	}
	pTiers := []cloudresourcesv1beta1.AlicloudRedisTier{
		cloudresourcesv1beta1.AlicloudRedisTierP1,
		cloudresourcesv1beta1.AlicloudRedisTierP2,
		cloudresourcesv1beta1.AlicloudRedisTierP3,
		cloudresourcesv1beta1.AlicloudRedisTierP4,
		cloudresourcesv1beta1.AlicloudRedisTierP5,
	}

	for i := range sTiers {
		sClass, sCount, err := redisTierToInstanceClassAndReadOnlyCount(sTiers[i])
		assert.NoError(t, err)
		pClass, pCount, err := redisTierToInstanceClassAndReadOnlyCount(pTiers[i])
		assert.NoError(t, err)
		assert.Contains(t, sClass, "redis.master.", "S%d should use redis.master.* cloud-disk class", i+1)
		assert.Contains(t, pClass, "redis.master.", "P%d should use redis.master.* cloud-disk class", i+1)
		assert.Equal(t, int32(0), sCount, "S%d should have readOnlyCount=0", i+1)
		assert.Equal(t, int32(1), pCount, "P%d should have readOnlyCount=1 (read replica)", i+1)
	}
}
