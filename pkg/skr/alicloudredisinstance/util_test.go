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
		{cloudresourcesv1beta1.AlicloudRedisTierS1, "tair.rdb.1g", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS2, "tair.rdb.2g", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS3, "tair.rdb.4g", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS4, "tair.rdb.8g", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierS5, "tair.rdb.16g", 0, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP1, "tair.rdb.4g", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP2, "tair.rdb.8g", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP3, "tair.rdb.16g", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP4, "tair.rdb.32g", 1, false},
		{cloudresourcesv1beta1.AlicloudRedisTierP5, "tair.rdb.64g", 1, false},
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
		assert.Contains(t, sClass, "tair.rdb.", "S%d should use tair.rdb.* cloud-disk class", i+1)
		assert.Contains(t, pClass, "tair.rdb.", "P%d should use tair.rdb.* cloud-disk class", i+1)
		assert.Equal(t, int32(0), sCount, "S%d should have readOnlyCount=0", i+1)
		assert.Equal(t, int32(1), pCount, "P%d should have readOnlyCount=1 (read replica)", i+1)
	}
}
