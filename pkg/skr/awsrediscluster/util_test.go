package awsrediscluster

import (
	"fmt"
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
)

type converterTestCase struct {
	InputRedisTier        cloudresourcesv1beta1.AwsRedisClusterTier
	ExpectedCacheNodeType string
}

func TestUtil(t *testing.T) {

	t.Run("redisTierToCacheNodeTypeConvertor", func(t *testing.T) {

		testCases := []converterTestCase{
			{cloudresourcesv1beta1.AwsRedisTierC1, "cache.t4g.small"},
			{cloudresourcesv1beta1.AwsRedisTierC2, "cache.t4g.medium"},
			{cloudresourcesv1beta1.AwsRedisTierC3, "cache.m7g.large"},
			{cloudresourcesv1beta1.AwsRedisTierC4, "cache.m7g.xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierC5, "cache.m7g.2xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierC6, "cache.m7g.4xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierC7, "cache.m7g.8xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierC8, "cache.m7g.16xlarge"},
		}

		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("should return expected result for input (%s)", testCase.InputRedisTier), func(t *testing.T) {
				cacheNodeType, err := redisTierToCacheNodeTypeConvertor(testCase.InputRedisTier)

				assert.Equal(t, testCase.ExpectedCacheNodeType, cacheNodeType, "resulting cacheNodeType does not match expected cacheNodeType")
				assert.Nil(t, err, "expected nil error, got an error")
			})

		}
		t.Run("should return error for unknown input", func(t *testing.T) {
			cacheNodeType, err := redisTierToCacheNodeTypeConvertor("unknown")

			assert.NotNil(t, err, "expected defined error, got nil")
			assert.Equal(t, "", cacheNodeType, "expected cacheNodeType to have zero value")
		})
	})

	type replicaConverterTestCase struct {
		InputRedisTier       cloudresourcesv1beta1.AwsRedisClusterTier
		ExpectedReadReplicas int32
	}
}
