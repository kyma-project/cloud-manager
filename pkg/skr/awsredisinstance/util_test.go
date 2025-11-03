package awsredisinstance

import (
	"fmt"
	"testing"

	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
)

type converterTestCase struct {
	InputRedisTier        cloudresourcesv1beta1.AwsRedisTier
	ExpectedCacheNodeType string
}

func TestUtil(t *testing.T) {

	t.Run("redisTierToCacheNodeTypeConvertor", func(t *testing.T) {

		testCases := []converterTestCase{
			{cloudresourcesv1beta1.AwsRedisTierS1, "cache.t4g.small"},
			{cloudresourcesv1beta1.AwsRedisTierS2, "cache.t4g.medium"},
			{cloudresourcesv1beta1.AwsRedisTierS3, "cache.m7g.large"},
			{cloudresourcesv1beta1.AwsRedisTierS4, "cache.m7g.xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierS5, "cache.m7g.2xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierS6, "cache.m7g.4xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierS7, "cache.m7g.8xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierS8, "cache.m7g.16xlarge"},

			{cloudresourcesv1beta1.AwsRedisTierP1, "cache.m7g.large"},
			{cloudresourcesv1beta1.AwsRedisTierP2, "cache.m7g.xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierP3, "cache.m7g.2xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierP4, "cache.m7g.4xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierP5, "cache.m7g.8xlarge"},
			{cloudresourcesv1beta1.AwsRedisTierP6, "cache.m7g.16xlarge"},
		}

		overwritesFromConfig := map[cloudresourcesv1beta1.AwsRedisTier]string{
			cloudresourcesv1beta1.AwsRedisTierS1: "custom.99.small",
			cloudresourcesv1beta1.AwsRedisTierS2: "custom.99.medium",
			cloudresourcesv1beta1.AwsRedisTierS3: "custom.99.large",
			cloudresourcesv1beta1.AwsRedisTierS4: "custom.99.xlarge",
			cloudresourcesv1beta1.AwsRedisTierS5: "custom.99.2xlarge",
			cloudresourcesv1beta1.AwsRedisTierS6: "custom.99.4xlarge",
			cloudresourcesv1beta1.AwsRedisTierS7: "custom.99.8xlarge",
			cloudresourcesv1beta1.AwsRedisTierS8: "custom.99.16xlarge",

			cloudresourcesv1beta1.AwsRedisTierP1: "custom.99.large",
			cloudresourcesv1beta1.AwsRedisTierP2: "custom.99.xlarge",
			cloudresourcesv1beta1.AwsRedisTierP3: "custom.99.2xlarge",
			cloudresourcesv1beta1.AwsRedisTierP4: "custom.99.4xlarge",
			cloudresourcesv1beta1.AwsRedisTierP5: "custom.99.8xlarge",
			cloudresourcesv1beta1.AwsRedisTierP6: "custom.99.16xlarge",
		}

		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("should return expected result for input (%s)", testCase.InputRedisTier), func(t *testing.T) {
				cacheNodeType, err := redisTierToCacheNodeTypeConvertor(testCase.InputRedisTier, nil)

				assert.Equal(t, testCase.ExpectedCacheNodeType, cacheNodeType, "resulting cacheNodeType does not match expected cacheNodeType")
				assert.Nil(t, err, "expected nil error, got an error")
			})

			t.Run(fmt.Sprintf("should return expected result for input (%s) (with overwrite config)", testCase.InputRedisTier), func(t *testing.T) {
				cacheNodeType, err := redisTierToCacheNodeTypeConvertor(testCase.InputRedisTier, overwritesFromConfig)

				assert.Equal(t, overwritesFromConfig[testCase.InputRedisTier], cacheNodeType, "resulting cacheNodeType does not match expected cacheNodeType")
				assert.Nil(t, err, "expected nil error, got an error")
			})
		}

		t.Run("should return error for unknown input", func(t *testing.T) {
			cacheNodeType, err := redisTierToCacheNodeTypeConvertor("unknown", nil)

			assert.NotNil(t, err, "expected defined error, got nil")
			assert.Equal(t, "", cacheNodeType, "expected cacheNodeType to have zero value")
		})

	})

	type replicaConverterTestCase struct {
		InputRedisTier       cloudresourcesv1beta1.AwsRedisTier
		ExpectedReadReplicas int32
	}

	t.Run("redisTierToReadReplicas", func(t *testing.T) {

		testCases := []replicaConverterTestCase{
			{cloudresourcesv1beta1.AwsRedisTierS1, int32(0)},
			{cloudresourcesv1beta1.AwsRedisTierS2, int32(0)},
			{cloudresourcesv1beta1.AwsRedisTierS3, int32(0)},
			{cloudresourcesv1beta1.AwsRedisTierS4, int32(0)},
			{cloudresourcesv1beta1.AwsRedisTierS5, int32(0)},
			{cloudresourcesv1beta1.AwsRedisTierS6, int32(0)},
			{cloudresourcesv1beta1.AwsRedisTierS7, int32(0)},
			{cloudresourcesv1beta1.AwsRedisTierS8, int32(0)},

			{cloudresourcesv1beta1.AwsRedisTierP1, int32(1)},
			{cloudresourcesv1beta1.AwsRedisTierP2, int32(1)},
			{cloudresourcesv1beta1.AwsRedisTierP3, int32(1)},
			{cloudresourcesv1beta1.AwsRedisTierP4, int32(1)},
			{cloudresourcesv1beta1.AwsRedisTierP5, int32(1)},
			{cloudresourcesv1beta1.AwsRedisTierP6, int32(1)},
		}

		for _, testCase := range testCases {
			t.Run(fmt.Sprintf("should return expected result for input (%s)", testCase.InputRedisTier), func(t *testing.T) {
				readReplicas := redisTierToReadReplicas(testCase.InputRedisTier)

				assert.Equal(t, testCase.ExpectedReadReplicas, readReplicas, "resulting readReplicas does not match expected readReplicas")
			})

		}
		t.Run("should return 0 for unknown input", func(t *testing.T) {
			readReplicas := redisTierToReadReplicas("unknown")

			assert.Equal(t, int32(0), readReplicas, "resulting readReplicas does not match expected readReplicas")
		})
	})
}
