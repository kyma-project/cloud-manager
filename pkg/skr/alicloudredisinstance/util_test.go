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
		{cloudresourcesv1beta1.AlicloudRedisTierS1, "redis.master.small.default", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS2, "redis.master.mid.default", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS3, "redis.master.stand.default", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS4, "redis.master.large.default", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS5, "redis.master.2xlarge.default", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP1, "redis.amber.master.small.multithread", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP2, "redis.amber.master.mid.multithread", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP3, "redis.amber.master.stand.multithread", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP4, "redis.amber.master.large.multithread", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP5, "redis.amber.master.2xlarge.multithread", 1, false},
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

func TestPTiersUseEnterpriseClass(t *testing.T) {
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
		assert.NotEqual(t, sClass, pClass, "S%d and P%d should use different instance classes", i+1, i+1)
		assert.Contains(t, sClass, "redis.master.", "S%d should use redis.master.* class", i+1)
		assert.Contains(t, pClass, "redis.amber.master.", "P%d should use redis.amber.master.* enterprise class", i+1)
		assert.Equal(t, int32(0), sCount, "S%d should have readOnlyCount=0", i+1)
		assert.Equal(t, int32(1), pCount, "P%d should have readOnlyCount=1 (enterprise read replica)", i+1)
	}
}
