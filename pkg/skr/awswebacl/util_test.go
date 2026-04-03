package awswebacl

import (
	"testing"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestConvertDefaultAction(t *testing.T) {
	t.Run("Allow action", func(t *testing.T) {
		result, err := convertDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultActionAllow)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Allow)
		assert.Nil(t, result.Block)
	})

	t.Run("Block action", func(t *testing.T) {
		result, err := convertDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultActionBlock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Allow)
		assert.NotNil(t, result.Block)
	})

	t.Run("Unknown action", func(t *testing.T) {
		result, err := convertDefaultAction("invalid")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestConvertRuleAction(t *testing.T) {
	t.Run("Allow action", func(t *testing.T) {
		result, err := convertRuleAction(cloudresourcesv1beta1.AwsWebAclRuleActionAllow)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Allow)
		assert.Nil(t, result.Block)
		assert.Nil(t, result.Count)
		assert.Nil(t, result.Captcha)
	})

	t.Run("Block action", func(t *testing.T) {
		result, err := convertRuleAction(cloudresourcesv1beta1.AwsWebAclRuleActionBlock)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Allow)
		assert.NotNil(t, result.Block)
		assert.Nil(t, result.Count)
		assert.Nil(t, result.Captcha)
	})

	t.Run("Count action", func(t *testing.T) {
		result, err := convertRuleAction(cloudresourcesv1beta1.AwsWebAclRuleActionCount)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Allow)
		assert.Nil(t, result.Block)
		assert.NotNil(t, result.Count)
		assert.Nil(t, result.Captcha)
	})

	t.Run("Captcha action", func(t *testing.T) {
		result, err := convertRuleAction(cloudresourcesv1beta1.AwsWebAclRuleActionCaptcha)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Allow)
		assert.Nil(t, result.Block)
		assert.Nil(t, result.Count)
		assert.NotNil(t, result.Captcha)
	})

	t.Run("Unknown action", func(t *testing.T) {
		result, err := convertRuleAction("invalid")
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestConvertVisibilityConfig(t *testing.T) {
	t.Run("With config", func(t *testing.T) {
		config := &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
			CloudWatchMetricsEnabled: true,
			MetricName:               "test-metric",
			SampledRequestsEnabled:   true,
		}
		result := convertVisibilityConfig(config, "default-name")
		assert.NotNil(t, result)
		assert.True(t, result.CloudWatchMetricsEnabled)
		assert.Equal(t, "test-metric", *result.MetricName)
		assert.True(t, result.SampledRequestsEnabled)
	})

	t.Run("With nil config uses defaults", func(t *testing.T) {
		result := convertVisibilityConfig(nil, "default-name")
		assert.NotNil(t, result)
		assert.False(t, result.CloudWatchMetricsEnabled)
		assert.Equal(t, "default-name", *result.MetricName)
		assert.False(t, result.SampledRequestsEnabled)
	})
}

func TestConvertGeoMatchStatement(t *testing.T) {
	t.Run("Multiple country codes", func(t *testing.T) {
		geo := &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
			CountryCodes: []string{"US", "GB", "DE"},
		}
		result := convertGeoMatchStatement(geo)
		assert.NotNil(t, result)
		assert.Len(t, result.CountryCodes, 3)
		assert.Equal(t, wafv2types.CountryCodeUs, result.CountryCodes[0])
		assert.Equal(t, wafv2types.CountryCodeGb, result.CountryCodes[1])
		assert.Equal(t, wafv2types.CountryCodeDe, result.CountryCodes[2])
	})

	t.Run("Empty country codes", func(t *testing.T) {
		geo := &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
			CountryCodes: []string{},
		}
		result := convertGeoMatchStatement(geo)
		assert.NotNil(t, result)
		assert.Len(t, result.CountryCodes, 0)
	})
}

func TestConvertManagedRuleGroupStatement(t *testing.T) {
	t.Run("With version and excluded rules", func(t *testing.T) {
		managed := &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
			VendorName: "AWS",
			Name:       "AWSManagedRulesCommonRuleSet",
			Version:    "Version_1.0",
			ExcludedRules: []cloudresourcesv1beta1.AwsWebAclExcludedRule{
				{Name: "SizeRestrictions_BODY"},
				{Name: "GenericRFI_BODY"},
			},
		}
		result := convertManagedRuleGroupStatement(managed)
		assert.NotNil(t, result)
		assert.Equal(t, "AWS", *result.VendorName)
		assert.Equal(t, "AWSManagedRulesCommonRuleSet", *result.Name)
		assert.Equal(t, "Version_1.0", *result.Version)
		assert.Len(t, result.ExcludedRules, 2)
		assert.Equal(t, "SizeRestrictions_BODY", *result.ExcludedRules[0].Name)
		assert.Equal(t, "GenericRFI_BODY", *result.ExcludedRules[1].Name)
	})

	t.Run("Without version and excluded rules", func(t *testing.T) {
		managed := &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
			VendorName: "AWS",
			Name:       "AWSManagedRulesBotControlRuleSet",
		}
		result := convertManagedRuleGroupStatement(managed)
		assert.NotNil(t, result)
		assert.Equal(t, "AWS", *result.VendorName)
		assert.Equal(t, "AWSManagedRulesBotControlRuleSet", *result.Name)
		assert.Nil(t, result.Version)
		assert.Len(t, result.ExcludedRules, 0)
	})
}

func TestConvertStatement(t *testing.T) {
	t.Run("GeoMatch statement", func(t *testing.T) {
		stmt := cloudresourcesv1beta1.AwsWebAclRuleStatement{
			GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
				CountryCodes: []string{"US"},
			},
		}
		result, err := convertStatement(stmt)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.GeoMatchStatement)
		assert.Nil(t, result.IPSetReferenceStatement)
		assert.Nil(t, result.RateBasedStatement)
		assert.Nil(t, result.ManagedRuleGroupStatement)
	})

	t.Run("ManagedRuleGroup statement", func(t *testing.T) {
		stmt := cloudresourcesv1beta1.AwsWebAclRuleStatement{
			ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
				VendorName: "AWS",
				Name:       "AWSManagedRulesCommonRuleSet",
			},
		}
		result, err := convertStatement(stmt)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ManagedRuleGroupStatement)
		assert.Nil(t, result.GeoMatchStatement)
		assert.Nil(t, result.IPSetReferenceStatement)
		assert.Nil(t, result.RateBasedStatement)
	})

	t.Run("Multiple statements returns error", func(t *testing.T) {
		stmt := cloudresourcesv1beta1.AwsWebAclRuleStatement{
			GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
				CountryCodes: []string{"US"},
			},
			ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
				VendorName: "AWS",
				Name:       "AWSManagedRulesCommonRuleSet",
			},
		}
		result, err := convertStatement(stmt)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "must have exactly one condition")
	})

	t.Run("No statements returns error", func(t *testing.T) {
		stmt := cloudresourcesv1beta1.AwsWebAclRuleStatement{}
		result, err := convertStatement(stmt)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "must have exactly one condition")
	})
}

func TestConvertRule(t *testing.T) {
	t.Run("Regular rule with action", func(t *testing.T) {
		rule := cloudresourcesv1beta1.AwsWebAclRule{
			Name:     "test-rule",
			Priority: 10,
			Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
			Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
				GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
					CountryCodes: []string{"US"},
				},
			},
			VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
				CloudWatchMetricsEnabled: true,
				MetricName:               "test-metric",
				SampledRequestsEnabled:   false,
			},
		}
		result, err := convertRule(rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "test-rule", *result.Name)
		assert.Equal(t, int32(10), result.Priority)
		assert.NotNil(t, result.Action)
		assert.NotNil(t, result.Action.Block)
		assert.Nil(t, result.OverrideAction)
	})

	t.Run("Managed rule group with override action", func(t *testing.T) {
		rule := cloudresourcesv1beta1.AwsWebAclRule{
			Name:     "managed-rule",
			Priority: 5,
			Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
				ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
					VendorName: "AWS",
					Name:       "AWSManagedRulesCommonRuleSet",
				},
			},
		}
		result, err := convertRule(rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "managed-rule", *result.Name)
		assert.Nil(t, result.Action)
		assert.NotNil(t, result.OverrideAction)
		assert.NotNil(t, result.OverrideAction.None)
	})
}

func TestConvertRules(t *testing.T) {
	t.Run("Multiple rules", func(t *testing.T) {
		rules := []cloudresourcesv1beta1.AwsWebAclRule{
			{
				Name:     "rule-1",
				Priority: 1,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionAllow,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
					},
				},
			},
			{
				Name:     "rule-2",
				Priority: 2,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
						VendorName: "AWS",
						Name:       "AWSManagedRulesCommonRuleSet",
					},
				},
			},
		}
		result, err := convertRules(rules)
		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, "rule-1", *result[0].Name)
		assert.Equal(t, "rule-2", *result[1].Name)
	})

	t.Run("Empty rules", func(t *testing.T) {
		result, err := convertRules([]cloudresourcesv1beta1.AwsWebAclRule{})
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Invalid rule returns error", func(t *testing.T) {
		rules := []cloudresourcesv1beta1.AwsWebAclRule{
			{
				Name:      "invalid-rule",
				Priority:  1,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{}, // No statement
			},
		}
		result, err := convertRules(rules)
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "invalid-rule")
	})
}
