package gcpredisinstance

import (
	"fmt"
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
)

type converterTestCase struct {
	InputRedisTier       cloudresourcesv1beta1.GcpRedisTier
	ExpectedTier         string
	ExpectedMemorySizeGb int32
}

func TestRedisTierToTierAndMemorySizeConverter(t *testing.T) {

	t.Run("redisTierToTierAndMemorySizeConverter", func(t *testing.T) {

		testCases := []converterTestCase{
			{cloudresourcesv1beta1.GcpRedisTierS1, "BASIC", 1},
			{cloudresourcesv1beta1.GcpRedisTierS2, "BASIC", 3},
			{cloudresourcesv1beta1.GcpRedisTierS3, "BASIC", 6},
			{cloudresourcesv1beta1.GcpRedisTierS4, "BASIC", 12},
			{cloudresourcesv1beta1.GcpRedisTierS5, "BASIC", 24},
			{cloudresourcesv1beta1.GcpRedisTierS6, "BASIC", 48},
			{cloudresourcesv1beta1.GcpRedisTierS7, "BASIC", 101},
			{cloudresourcesv1beta1.GcpRedisTierS8, "BASIC", 200},

			{cloudresourcesv1beta1.GcpRedisTierP1, "STANDARD_HA", 5},
			{cloudresourcesv1beta1.GcpRedisTierP2, "STANDARD_HA", 12},
			{cloudresourcesv1beta1.GcpRedisTierP3, "STANDARD_HA", 24},
			{cloudresourcesv1beta1.GcpRedisTierP4, "STANDARD_HA", 48},
			{cloudresourcesv1beta1.GcpRedisTierP5, "STANDARD_HA", 101},
			{cloudresourcesv1beta1.GcpRedisTierP6, "STANDARD_HA", 200},
		}

		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("should return expected result for input (%s)", testCase.InputRedisTier), func(t *testing.T) {
				resultTier, resultMemorySizeGb, err := redisTierToTierAndMemorySizeConverter(testCase.InputRedisTier)

				assert.Equal(t, testCase.ExpectedTier, resultTier, "resulting tier does not match expected tier")
				assert.Equal(t, testCase.ExpectedMemorySizeGb, resultMemorySizeGb, "resulting memorySizeGb does not match expected memorySizeGb")
				assert.Nil(t, err, "expected nil error, got an error")
			})

		}
		t.Run("should return error for unknown input", func(t *testing.T) {
			resultTier, resultMemorySizeGb, err := redisTierToTierAndMemorySizeConverter("unknown")

			assert.NotNil(t, err, "expected defined error, got nil")
			assert.Equal(t, "", resultTier, "expected tier to have zero value")
			assert.Equal(t, int32(0), resultMemorySizeGb, "expected memorySizeGb to have zero value")
		})
	})

	type replicaConverterTestCase struct {
		InputRedisTier       cloudresourcesv1beta1.GcpRedisTier
		ExpectedReadReplicas int32
	}

	t.Run("redisTierToReadReplicas", func(t *testing.T) {

		testCases := []replicaConverterTestCase{
			{cloudresourcesv1beta1.GcpRedisTierS1, int32(0)},
			{cloudresourcesv1beta1.GcpRedisTierS2, int32(0)},
			{cloudresourcesv1beta1.GcpRedisTierS3, int32(0)},
			{cloudresourcesv1beta1.GcpRedisTierS4, int32(0)},
			{cloudresourcesv1beta1.GcpRedisTierS5, int32(0)},
			{cloudresourcesv1beta1.GcpRedisTierS6, int32(0)},
			{cloudresourcesv1beta1.GcpRedisTierS7, int32(0)},
			{cloudresourcesv1beta1.GcpRedisTierS8, int32(0)},

			{cloudresourcesv1beta1.GcpRedisTierP1, int32(1)},
			{cloudresourcesv1beta1.GcpRedisTierP2, int32(1)},
			{cloudresourcesv1beta1.GcpRedisTierP3, int32(1)},
			{cloudresourcesv1beta1.GcpRedisTierP4, int32(1)},
			{cloudresourcesv1beta1.GcpRedisTierP5, int32(1)},
			{cloudresourcesv1beta1.GcpRedisTierP6, int32(1)},
		}

		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("should return expected result for input (%s)", testCase.InputRedisTier), func(t *testing.T) {
				readReplicas := redisTierToReplicaCount(testCase.InputRedisTier)

				assert.Equal(t, testCase.ExpectedReadReplicas, readReplicas, "resulting readReplicas does not match expected readReplicas")
			})
		}

		t.Run("should return 0 for unknown input", func(t *testing.T) {
			readReplicas := redisTierToReplicaCount("unknown")

			assert.Equal(t, int32(0), readReplicas, "resulting readReplicas does not match expected readReplicas")
		})
	})
}
