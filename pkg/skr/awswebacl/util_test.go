package awswebacl

import (
	"testing"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestConvertDefaultAction(t *testing.T) {
	t.Run("Allow action", func(t *testing.T) {
		action := cloudresourcesv1beta1.AwsWebAclDefaultAction{
			Allow: &cloudresourcesv1beta1.AwsWebAclAllowAction{},
		}

		result, err := convertDefaultAction(action)
		require.NoError(t, err)
		assert.NotNil(t, result.Allow)
		assert.Nil(t, result.Block)
	})

	t.Run("Block action", func(t *testing.T) {
		action := cloudresourcesv1beta1.AwsWebAclDefaultAction{
			Block: &cloudresourcesv1beta1.AwsWebAclBlockAction{},
		}

		result, err := convertDefaultAction(action)
		require.NoError(t, err)
		assert.NotNil(t, result.Block)
		assert.Nil(t, result.Allow)
	})
}

func TestConvertVisibilityConfig(t *testing.T) {
	t.Run("With full config", func(t *testing.T) {
		config := &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               "test-metric",
			SampledRequestsEnabled:   true,
		}

		result := convertVisibilityConfig(config, "default-name")
		assert.True(t, result.CloudWatchMetricsEnabled)
		assert.Equal(t, "test-metric", *result.MetricName)
		assert.True(t, result.SampledRequestsEnabled)
	})

	t.Run("With nil config uses default name", func(t *testing.T) {
		result := convertVisibilityConfig(nil, "default-name")
		assert.False(t, result.CloudWatchMetricsEnabled)
		assert.Equal(t, "default-name", *result.MetricName)
		assert.False(t, result.SampledRequestsEnabled)
	})
}

func TestConvertManagedRuleGroupStatement(t *testing.T) {
	t.Run("Basic managed rule group", func(t *testing.T) {
		managed := &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
			VendorName: "AWS",
			Name:       "AWSManagedRulesCommonRuleSet",
		}

		result, err := convertManagedRuleGroupStatement(managed)
		require.NoError(t, err)
		assert.Equal(t, "AWS", *result.VendorName)
		assert.Equal(t, "AWSManagedRulesCommonRuleSet", *result.Name)
		assert.Nil(t, result.Version)
	})

	t.Run("With version", func(t *testing.T) {
		managed := &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
			VendorName: "AWS",
			Name:       "AWSManagedRulesCommonRuleSet",
			Version:    "1.0",
		}

		result, err := convertManagedRuleGroupStatement(managed)
		require.NoError(t, err)
		assert.Equal(t, "1.0", *result.Version)
	})

	t.Run("With excluded rules", func(t *testing.T) {
		managed := &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
			VendorName: "AWS",
			Name:       "AWSManagedRulesCommonRuleSet",
			ExcludedRules: []cloudresourcesv1beta1.AwsWebAclExcludedRule{
				{Name: "SizeRestrictions_BODY"},
				{Name: "GenericRFI_BODY"},
			},
		}

		result, err := convertManagedRuleGroupStatement(managed)
		require.NoError(t, err)
		assert.Len(t, result.ExcludedRules, 2)
		assert.Equal(t, "SizeRestrictions_BODY", *result.ExcludedRules[0].Name)
		assert.Equal(t, "GenericRFI_BODY", *result.ExcludedRules[1].Name)
	})

	t.Run("With action overrides", func(t *testing.T) {
		managed := &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
			VendorName: "AWS",
			Name:       "AWSManagedRulesCommonRuleSet",
			RuleActionOverrides: []cloudresourcesv1beta1.AwsWebAclRuleActionOverride{
				{
					Name: "SizeRestrictions_BODY",
					ActionToUse: &cloudresourcesv1beta1.AwsWebAclRuleAction{
						Count: &cloudresourcesv1beta1.AwsWebAclCountAction{},
					},
				},
			},
		}

		result, err := convertManagedRuleGroupStatement(managed)
		require.NoError(t, err)
		assert.Len(t, result.RuleActionOverrides, 1)
		assert.Equal(t, "SizeRestrictions_BODY", *result.RuleActionOverrides[0].Name)
		assert.NotNil(t, result.RuleActionOverrides[0].ActionToUse.Count)
	})
}

func TestConvertRule(t *testing.T) {
	t.Run("Rule with ManagedRuleGroup and OverrideAction", func(t *testing.T) {
		rule := cloudresourcesv1beta1.AwsWebAclRule{
			Name:           "managed-rule",
			Priority:       1,
			OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
			Statements: []cloudresourcesv1beta1.AwsWebAclStatement{
				{
					ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
						VendorName: "AWS",
						Name:       "AWSManagedRulesCommonRuleSet",
					},
				},
			},
		}

		result, err := convertRule(rule)
		require.NoError(t, err)
		assert.Equal(t, "managed-rule", *result.Name)
		assert.Equal(t, int32(1), result.Priority)
		assert.NotNil(t, result.OverrideAction)
		assert.NotNil(t, result.OverrideAction.None)
	})

	t.Run("Rule without OverrideAction defaults to None", func(t *testing.T) {
		rule := cloudresourcesv1beta1.AwsWebAclRule{
			Name:     "managed-rule-default",
			Priority: 2,
			// OverrideAction not specified - should default to None
			Statements: []cloudresourcesv1beta1.AwsWebAclStatement{
				{
					ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
						VendorName: "AWS",
						Name:       "AWSManagedRulesKnownBadInputsRuleSet",
					},
				},
			},
		}

		result, err := convertRule(rule)
		require.NoError(t, err)
		assert.Equal(t, "managed-rule-default", *result.Name)
		assert.Equal(t, int32(2), result.Priority)
		assert.NotNil(t, result.OverrideAction, "OverrideAction should default to None when not specified")
		assert.NotNil(t, result.OverrideAction.None, "OverrideAction.None should be set as default")
	})

	t.Run("Rule with empty OverrideAction defaults to None", func(t *testing.T) {
		rule := cloudresourcesv1beta1.AwsWebAclRule{
			Name:     "managed-rule-empty",
			Priority: 3,
			// OverrideAction specified but empty - should default to None
			OverrideAction: &cloudresourcesv1beta1.AwsWebAclOverrideAction{},
			Statements: []cloudresourcesv1beta1.AwsWebAclStatement{
				{
					ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
						VendorName: "AWS",
						Name:       "AWSManagedRulesSQLiRuleSet",
					},
				},
			},
		}

		result, err := convertRule(rule)
		require.NoError(t, err)
		assert.Equal(t, "managed-rule-empty", *result.Name)
		assert.Equal(t, int32(3), result.Priority)
		assert.NotNil(t, result.OverrideAction, "OverrideAction should default to None when empty")
		assert.NotNil(t, result.OverrideAction.None, "OverrideAction.None should be set as default")
	})
}

func TestConvertRules(t *testing.T) {
	rules := []cloudresourcesv1beta1.AwsWebAclRule{
		{
			Name:           "rule-1",
			Priority:       0,
			OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
			Statements: []cloudresourcesv1beta1.AwsWebAclStatement{
				{
					ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
						VendorName: "AWS",
						Name:       "AWSManagedRulesCommonRuleSet",
					},
				},
			},
		},
		{
			Name:           "rule-2",
			Priority:       1,
			OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
			Statements: []cloudresourcesv1beta1.AwsWebAclStatement{
				{
					ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
						VendorName: "AWS",
						Name:       "AWSManagedRulesSQLiRuleSet",
					},
				},
			},
		},
	}

	result, err := convertRules(rules)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "rule-1", *result[0].Name)
	assert.Equal(t, "rule-2", *result[1].Name)
}

func TestConvertRuleActionType(t *testing.T) {
	t.Run("Challenge action", func(t *testing.T) {
		action := &cloudresourcesv1beta1.AwsWebAclRuleAction{
			Challenge: &cloudresourcesv1beta1.AwsWebAclChallengeAction{},
		}

		result, err := convertRuleActionType(action)
		require.NoError(t, err)
		assert.NotNil(t, result.Challenge)
	})

	t.Run("Captcha action", func(t *testing.T) {
		action := &cloudresourcesv1beta1.AwsWebAclRuleAction{
			Captcha: &cloudresourcesv1beta1.AwsWebAclCaptchaAction{},
		}

		result, err := convertRuleActionType(action)
		require.NoError(t, err)
		assert.NotNil(t, result.Captcha)
	})
}

func TestConvertTags(t *testing.T) {
	webAcl := &cloudresourcesv1beta1.AwsWebAcl{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-acl",
		},
	}

	scope := &cloudcontrolv1beta1.Scope{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-scope",
		},
		Spec: cloudcontrolv1beta1.ScopeSpec{
			ShootName: "test-shoot",
		},
	}

	result := convertTags(webAcl, scope)
	assert.Len(t, result, 4)
	assert.Equal(t, "Name", *result[0].Key)
	assert.Equal(t, "test-acl", *result[0].Value)
	assert.Equal(t, "ManagedBy", *result[1].Key)
	assert.Equal(t, "cloud-manager", *result[1].Value)
}

func TestConvertCaptchaConfig(t *testing.T) {
	config := &cloudresourcesv1beta1.AwsWebAclCaptchaConfig{
		ImmunityTime: 300,
	}

	result := convertCaptchaConfig(config)
	require.NotNil(t, result)
	assert.Equal(t, int64(300), *result.ImmunityTimeProperty.ImmunityTime)
}

func TestConvertChallengeConfig(t *testing.T) {
	config := &cloudresourcesv1beta1.AwsWebAclChallengeConfig{
		ImmunityTime: 600,
	}

	result := convertChallengeConfig(config)
	require.NotNil(t, result)
	assert.Equal(t, int64(600), *result.ImmunityTimeProperty.ImmunityTime)
}

func TestConvertCustomRequestHandling(t *testing.T) {
	handling := &cloudresourcesv1beta1.AwsWebAclCustomRequestHandling{
		InsertHeaders: []cloudresourcesv1beta1.AwsWebAclCustomHTTPHeader{
			{Name: "X-Custom-Header", Value: "custom-value"},
		},
	}

	result := convertCustomRequestHandling(handling)
	require.NotNil(t, result)
	assert.Len(t, result.InsertHeaders, 1)
	assert.Equal(t, "X-Custom-Header", *result.InsertHeaders[0].Name)
	assert.Equal(t, "custom-value", *result.InsertHeaders[0].Value)
}

func TestConvertCustomResponse(t *testing.T) {
	response := &cloudresourcesv1beta1.AwsWebAclCustomResponse{
		ResponseCode:          403,
		CustomResponseBodyKey: "blocked",
		ResponseHeaders: []cloudresourcesv1beta1.AwsWebAclCustomHTTPHeader{
			{Name: "X-Error", Value: "Blocked"},
		},
	}

	result := convertCustomResponse(response)
	require.NotNil(t, result)
	assert.Equal(t, int32(403), *result.ResponseCode)
	assert.Equal(t, "blocked", *result.CustomResponseBodyKey)
	assert.Len(t, result.ResponseHeaders, 1)
}

func TestConvertCustomResponseBodies(t *testing.T) {
	bodies := map[string]cloudresourcesv1beta1.AwsWebAclCustomResponseBody{
		"blocked": {
			ContentType: "TEXT_PLAIN",
			Content:     "Access Denied",
		},
	}

	result := convertCustomResponseBodies(bodies)
	require.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, wafv2types.ResponseContentTypeTextPlain, result["blocked"].ContentType)
	assert.Equal(t, "Access Denied", *result["blocked"].Content)
}

func TestScopeRegional(t *testing.T) {
	result := ScopeRegional()
	assert.Equal(t, wafv2types.ScopeRegional, result)
}

func TestConvertOverrideAction(t *testing.T) {
	t.Run("None override", func(t *testing.T) {
		override := &cloudresourcesv1beta1.AwsWebAclOverrideAction{
			None: &cloudresourcesv1beta1.AwsWebAclNoneAction{},
		}

		result, err := convertOverrideAction(override)
		require.NoError(t, err)
		assert.NotNil(t, result.None)
		assert.Nil(t, result.Count)
	})

	t.Run("Count override", func(t *testing.T) {
		override := &cloudresourcesv1beta1.AwsWebAclOverrideAction{
			Count: &cloudresourcesv1beta1.AwsWebAclCountAction{},
		}

		result, err := convertOverrideAction(override)
		require.NoError(t, err)
		assert.NotNil(t, result.Count)
		assert.Nil(t, result.None)
	})
}
