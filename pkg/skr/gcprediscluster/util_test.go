package gcprediscluster

import (
	"fmt"
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
)

type converterTestCase struct {
	InputRedisTier   cloudresourcesv1beta1.GcpRedisClusterTier
	ExpectedNodeType string
}

func TestRedisTierToNodeTypeConverter(t *testing.T) {

	t.Run("redisTierToNodeTypeConverter", func(t *testing.T) {

		testCases := []converterTestCase{
			{cloudresourcesv1beta1.GcpRedisClusterTierC1, "REDIS_SHARED_CORE_NANO"},
			{cloudresourcesv1beta1.GcpRedisClusterTierC3, "REDIS_STANDARD_SMALL"},
			{cloudresourcesv1beta1.GcpRedisClusterTierC4, "REDIS_HIGHMEM_MEDIUM"},
			{cloudresourcesv1beta1.GcpRedisClusterTierC6, "REDIS_HIGHMEM_XLARGE"},
		}

		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("should return expected result for input (%s)", testCase.InputRedisTier), func(t *testing.T) {
				resultNodeType, err := redisTierToNodeTypeConverter(testCase.InputRedisTier)

				assert.Equal(t, testCase.ExpectedNodeType, resultNodeType, "resulting tier does not match expected node type")
				assert.Nil(t, err, "expected nil error, got an error")
			})

		}
		t.Run("should return error for unknown input", func(t *testing.T) {
			resultNodeType, err := redisTierToNodeTypeConverter("unknown")

			assert.NotNil(t, err, "expected defined error, got nil")
			assert.Equal(t, "", resultNodeType, "expected nodeType to have zero value")
		})
	})

	type replicaConverterTestCase struct {
		InputRedisTier       cloudresourcesv1beta1.GcpRedisTier
		ExpectedReadReplicas int32
	}

}
