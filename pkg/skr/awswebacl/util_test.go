package awswebacl

import (
	"testing"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/stretchr/testify/assert"
)

func TestConvertDefaultAction(t *testing.T) {
	t.Run("Allow action", func(t *testing.T) {
		result, err := convertDefaultAction(cloudresourcesv1beta1.DefaultActionAllow())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Allow)
		assert.Nil(t, result.Block)
	})

	t.Run("Block action", func(t *testing.T) {
		result, err := convertDefaultAction(cloudresourcesv1beta1.DefaultActionBlock())
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Allow)
		assert.NotNil(t, result.Block)
	})

	t.Run("Empty action returns error", func(t *testing.T) {
		result, err := convertDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultAction{}) // Neither Allow nor Block
		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "must have either allow or block")
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
		result, err := convertGeoMatchStatement(geo)
		assert.NoError(t, err)
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
		result, err := convertGeoMatchStatement(geo)
		assert.NoError(t, err)
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
		result, err := convertManagedRuleGroupStatement(managed)
		assert.NoError(t, err)
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
		result, err := convertManagedRuleGroupStatement(managed)
		assert.NoError(t, err)
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
			Action:   cloudresourcesv1beta1.RuleActionBlock(),
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
			Name:           "managed-rule",
			Priority:       5,
			OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
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
				Action:   cloudresourcesv1beta1.RuleActionAllow(),
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
					},
				},
			},
			{
				Name:           "rule-2",
				Priority:       2,
				OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
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

func TestConvertByteMatchStatement(t *testing.T) {
	t.Run("With all fields", func(t *testing.T) {
		byteMatch := &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
			SearchString:         "../",
			PositionalConstraint: "CONTAINS",
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				QueryString: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "NONE"},
			},
		}
		result, err := convertByteMatchStatement(byteMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, []byte("../"), result.SearchString)
		assert.Equal(t, wafv2types.PositionalConstraintContains, result.PositionalConstraint)
		assert.NotNil(t, result.FieldToMatch)
		assert.NotNil(t, result.FieldToMatch.QueryString)
		assert.Len(t, result.TextTransformations, 1)
		assert.Equal(t, int32(0), result.TextTransformations[0].Priority)
		assert.Equal(t, wafv2types.TextTransformationTypeNone, result.TextTransformations[0].Type)
	})

	t.Run("With multiple transformations", func(t *testing.T) {
		byteMatch := &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
			SearchString:         "admin",
			PositionalConstraint: "STARTS_WITH",
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				UriPath: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "LOWERCASE"},
				{Priority: 1, Type: "URL_DECODE"},
			},
		}
		result, err := convertByteMatchStatement(byteMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.TextTransformations, 2)
		assert.Equal(t, wafv2types.TextTransformationTypeLowercase, result.TextTransformations[0].Type)
		assert.Equal(t, wafv2types.TextTransformationTypeUrlDecode, result.TextTransformations[1].Type)
	})
}

func TestConvertFieldToMatch(t *testing.T) {
	t.Run("UriPath", func(t *testing.T) {
		field := cloudresourcesv1beta1.AwsWebAclFieldToMatch{UriPath: true}
		result, err := convertFieldToMatch(field)
		assert.NoError(t, err)
		assert.NotNil(t, result.UriPath)
		assert.Nil(t, result.QueryString)
	})

	t.Run("QueryString", func(t *testing.T) {
		field := cloudresourcesv1beta1.AwsWebAclFieldToMatch{QueryString: true}
		result, err := convertFieldToMatch(field)
		assert.NoError(t, err)
		assert.NotNil(t, result.QueryString)
		assert.Nil(t, result.UriPath)
	})

	t.Run("Method", func(t *testing.T) {
		field := cloudresourcesv1beta1.AwsWebAclFieldToMatch{Method: true}
		result, err := convertFieldToMatch(field)
		assert.NoError(t, err)
		assert.NotNil(t, result.Method)
	})

	t.Run("SingleHeader", func(t *testing.T) {
		field := cloudresourcesv1beta1.AwsWebAclFieldToMatch{SingleHeader: "user-agent"}
		result, err := convertFieldToMatch(field)
		assert.NoError(t, err)
		assert.NotNil(t, result.SingleHeader)
		assert.Equal(t, "user-agent", *result.SingleHeader.Name)
	})

	t.Run("Body", func(t *testing.T) {
		field := cloudresourcesv1beta1.AwsWebAclFieldToMatch{Body: true}
		result, err := convertFieldToMatch(field)
		assert.NoError(t, err)
		assert.NotNil(t, result.Body)
		assert.Equal(t, wafv2types.OversizeHandlingContinue, result.Body.OversizeHandling)
	})
}

func TestConvertTextTransformations(t *testing.T) {
	t.Run("Single transformation", func(t *testing.T) {
		transforms := []cloudresourcesv1beta1.AwsWebAclTextTransformation{
			{Priority: 0, Type: "NONE"},
		}
		result, err := convertTextTransformations(transforms)
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, int32(0), result[0].Priority)
		assert.Equal(t, wafv2types.TextTransformationTypeNone, result[0].Type)
	})

	t.Run("Multiple transformations", func(t *testing.T) {
		transforms := []cloudresourcesv1beta1.AwsWebAclTextTransformation{
			{Priority: 0, Type: "LOWERCASE"},
			{Priority: 1, Type: "URL_DECODE"},
			{Priority: 2, Type: "HTML_ENTITY_DECODE"},
		}
		result, err := convertTextTransformations(transforms)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
	})

	t.Run("Empty transformations returns error", func(t *testing.T) {
		result, err := convertTextTransformations([]cloudresourcesv1beta1.AwsWebAclTextTransformation{})
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestConvertForwardedIPConfig(t *testing.T) {
	t.Run("With all fields", func(t *testing.T) {
		config := &cloudresourcesv1beta1.AwsWebAclForwardedIPConfig{
			HeaderName:       "X-Forwarded-For",
			FallbackBehavior: "MATCH",
		}
		result := convertForwardedIPConfig(config)
		assert.NotNil(t, result)
		assert.Equal(t, "X-Forwarded-For", *result.HeaderName)
		assert.Equal(t, wafv2types.FallbackBehaviorMatch, result.FallbackBehavior)
	})
}

func TestConvertGeoMatchStatementWithForwardedIP(t *testing.T) {
	t.Run("With ForwardedIPConfig", func(t *testing.T) {
		geo := &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
			CountryCodes: []string{"US"},
			ForwardedIPConfig: &cloudresourcesv1beta1.AwsWebAclForwardedIPConfig{
				HeaderName:       "X-Forwarded-For",
				FallbackBehavior: "NO_MATCH",
			},
		}
		result, err := convertGeoMatchStatement(geo)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ForwardedIPConfig)
		assert.Equal(t, "X-Forwarded-For", *result.ForwardedIPConfig.HeaderName)
		assert.Equal(t, wafv2types.FallbackBehaviorNoMatch, result.ForwardedIPConfig.FallbackBehavior)
	})
}

func TestConvertRateBasedStatementWithForwardedIP(t *testing.T) {
	t.Run("With ForwardedIPConfig", func(t *testing.T) {
		rate := &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
			Limit: 2000,
			ForwardedIPConfig: &cloudresourcesv1beta1.AwsWebAclForwardedIPConfig{
				HeaderName:       "X-Forwarded-For",
				FallbackBehavior: "MATCH",
			},
		}
		result, err := convertRateBasedStatement(rate)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ForwardedIPConfig)
		assert.Equal(t, "X-Forwarded-For", *result.ForwardedIPConfig.HeaderName)
	})
}

func TestConvertManagedRuleGroupStatementWithConfigs(t *testing.T) {
	t.Run("With ManagedRuleGroupConfigs", func(t *testing.T) {
		managed := &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
			VendorName: "AWS",
			Name:       "AWSManagedRulesATPRuleSet",
			ManagedRuleGroupConfigs: []cloudresourcesv1beta1.AwsWebAclManagedRuleGroupConfig{
				{
					LoginPath:   "/login",
					PayloadType: "JSON",
					UsernameField: &cloudresourcesv1beta1.AwsWebAclFieldIdentifier{
						Identifier: "/username",
					},
					PasswordField: &cloudresourcesv1beta1.AwsWebAclFieldIdentifier{
						Identifier: "/password",
					},
				},
			},
		}
		result, err := convertManagedRuleGroupStatement(managed)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.ManagedRuleGroupConfigs, 1)
		assert.Equal(t, "/login", *result.ManagedRuleGroupConfigs[0].LoginPath)
		assert.Equal(t, wafv2types.PayloadTypeJson, result.ManagedRuleGroupConfigs[0].PayloadType)
		assert.Equal(t, "/username", *result.ManagedRuleGroupConfigs[0].UsernameField.Identifier)
		assert.Equal(t, "/password", *result.ManagedRuleGroupConfigs[0].PasswordField.Identifier)
	})

	t.Run("With RuleActionOverrides", func(t *testing.T) {
		managed := &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
			VendorName: "AWS",
			Name:       "AWSManagedRulesCommonRuleSet",
			RuleActionOverrides: []cloudresourcesv1beta1.AwsWebAclRuleActionOverride{
				{
					Name:        "SizeRestrictions_BODY",
					ActionToUse: cloudresourcesv1beta1.AwsWebAclRuleActionCount,
				},
				{
					Name:        "GenericRFI_BODY",
					ActionToUse: cloudresourcesv1beta1.AwsWebAclRuleActionAllow,
				},
			},
		}
		result, err := convertManagedRuleGroupStatement(managed)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.RuleActionOverrides, 2)
		assert.Equal(t, "SizeRestrictions_BODY", *result.RuleActionOverrides[0].Name)
		assert.NotNil(t, result.RuleActionOverrides[0].ActionToUse.Count)
		assert.Equal(t, "GenericRFI_BODY", *result.RuleActionOverrides[1].Name)
		assert.NotNil(t, result.RuleActionOverrides[1].ActionToUse.Allow)
	})
}

func TestConvertAssociationConfig(t *testing.T) {
	t.Run("With request body config", func(t *testing.T) {
		config := &cloudresourcesv1beta1.AwsWebAclAssociationConfig{
			RequestBody: map[string]cloudresourcesv1beta1.AwsWebAclRequestBodyConfig{
				"CLOUDFRONT": {
					DefaultSizeInspectionLimit: "KB_32",
				},
				"API_GATEWAY": {
					DefaultSizeInspectionLimit: "KB_16",
				},
			},
		}
		result := convertAssociationConfig(config)
		assert.NotNil(t, result)
		assert.Len(t, result.RequestBody, 2)
		assert.Equal(t, wafv2types.SizeInspectionLimitKb32, result.RequestBody["CLOUDFRONT"].DefaultSizeInspectionLimit)
		assert.Equal(t, wafv2types.SizeInspectionLimitKb16, result.RequestBody["API_GATEWAY"].DefaultSizeInspectionLimit)
	})

	t.Run("With nil config returns nil", func(t *testing.T) {
		result := convertAssociationConfig(nil)
		assert.Nil(t, result)
	})

	t.Run("With empty RequestBody returns nil", func(t *testing.T) {
		config := &cloudresourcesv1beta1.AwsWebAclAssociationConfig{
			RequestBody: map[string]cloudresourcesv1beta1.AwsWebAclRequestBodyConfig{},
		}
		result := convertAssociationConfig(config)
		assert.Nil(t, result)
	})
}

func TestConvertRuleLabels(t *testing.T) {
	t.Run("With multiple labels", func(t *testing.T) {
		labels := []cloudresourcesv1beta1.AwsWebAclLabel{
			{Name: "label:one"},
			{Name: "label:two"},
			{Name: "namespace:production"},
		}
		result := convertRuleLabels(labels)
		assert.NotNil(t, result)
		assert.Len(t, result, 3)
		assert.Equal(t, "label:one", *result[0].Name)
		assert.Equal(t, "label:two", *result[1].Name)
		assert.Equal(t, "namespace:production", *result[2].Name)
	})

	t.Run("With empty labels returns nil", func(t *testing.T) {
		result := convertRuleLabels([]cloudresourcesv1beta1.AwsWebAclLabel{})
		assert.Nil(t, result)
	})
}

func TestConvertRuleActionTypeWithChallenge(t *testing.T) {
	t.Run("Challenge action", func(t *testing.T) {
		actionType := &cloudresourcesv1beta1.AwsWebAclRuleActionType{
			Challenge: &cloudresourcesv1beta1.AwsWebAclChallengeAction{},
		}
		result, err := convertRuleActionType(actionType)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.Allow)
		assert.Nil(t, result.Block)
		assert.Nil(t, result.Count)
		assert.Nil(t, result.Captcha)
		assert.NotNil(t, result.Challenge)
	})

	t.Run("Challenge action with custom request handling", func(t *testing.T) {
		actionType := &cloudresourcesv1beta1.AwsWebAclRuleActionType{
			Challenge: &cloudresourcesv1beta1.AwsWebAclChallengeAction{
				CustomRequestHandling: &cloudresourcesv1beta1.AwsWebAclCustomRequestHandling{
					InsertHeaders: []cloudresourcesv1beta1.AwsWebAclCustomHTTPHeader{
						{Name: "X-Challenge-Passed", Value: "true"},
					},
				},
			},
		}
		result, err := convertRuleActionType(actionType)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Challenge)
		assert.NotNil(t, result.Challenge.CustomRequestHandling)
		assert.Len(t, result.Challenge.CustomRequestHandling.InsertHeaders, 1)
		assert.Equal(t, "X-Challenge-Passed", *result.Challenge.CustomRequestHandling.InsertHeaders[0].Name)
	})
}

func TestConvertRuleWithPerRuleConfigs(t *testing.T) {
	t.Run("Rule with per-rule CaptchaConfig overrides global", func(t *testing.T) {
		rule := cloudresourcesv1beta1.AwsWebAclRule{
			Name:     "captcha-rule-with-override",
			Priority: 10,
			Action:   cloudresourcesv1beta1.RuleActionCaptcha(),
			Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
				GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
					CountryCodes: []string{"US"},
				},
			},
			CaptchaConfig: &cloudresourcesv1beta1.AwsWebAclCaptchaConfig{
				ImmunityTime: 300, // 5 minutes override
			},
		}
		result, err := convertRule(rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.CaptchaConfig)
		assert.Equal(t, int64(300), *result.CaptchaConfig.ImmunityTimeProperty.ImmunityTime)
	})

	t.Run("Rule with per-rule ChallengeConfig overrides global", func(t *testing.T) {
		rule := cloudresourcesv1beta1.AwsWebAclRule{
			Name:     "challenge-rule-with-override",
			Priority: 20,
			Action:   cloudresourcesv1beta1.RuleActionChallenge(),
			Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
				GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
					CountryCodes: []string{"CN"},
				},
			},
			ChallengeConfig: &cloudresourcesv1beta1.AwsWebAclChallengeConfig{
				ImmunityTime: 600, // 10 minutes override
			},
		}
		result, err := convertRule(rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.ChallengeConfig)
		assert.Equal(t, int64(600), *result.ChallengeConfig.ImmunityTimeProperty.ImmunityTime)
	})

	t.Run("Rule with both per-rule configs", func(t *testing.T) {
		rule := cloudresourcesv1beta1.AwsWebAclRule{
			Name:     "rule-with-both-configs",
			Priority: 30,
			Action:   cloudresourcesv1beta1.RuleActionCount(),
			Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
				RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
					Limit: 2000,
				},
			},
			CaptchaConfig: &cloudresourcesv1beta1.AwsWebAclCaptchaConfig{
				ImmunityTime: 300,
			},
			ChallengeConfig: &cloudresourcesv1beta1.AwsWebAclChallengeConfig{
				ImmunityTime: 600,
			},
			RuleLabels: []cloudresourcesv1beta1.AwsWebAclLabel{
				{Name: "rate-limited"},
			},
		}
		result, err := convertRule(rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.CaptchaConfig)
		assert.NotNil(t, result.ChallengeConfig)
		assert.Len(t, result.RuleLabels, 1)
		assert.Equal(t, int64(300), *result.CaptchaConfig.ImmunityTimeProperty.ImmunityTime)
		assert.Equal(t, int64(600), *result.ChallengeConfig.ImmunityTimeProperty.ImmunityTime)
	})

	t.Run("Rule without per-rule configs has nil", func(t *testing.T) {
		rule := cloudresourcesv1beta1.AwsWebAclRule{
			Name:     "simple-rule",
			Priority: 40,
			Action:   cloudresourcesv1beta1.RuleActionAllow(),
			Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
				GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
					CountryCodes: []string{"US"},
				},
			},
		}
		result, err := convertRule(rule)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Nil(t, result.CaptchaConfig)
		assert.Nil(t, result.ChallengeConfig)
	})
}

func TestConvertLabelMatchStatement(t *testing.T) {
	t.Run("LabelMatch with LABEL scope", func(t *testing.T) {
		labelMatch := &cloudresourcesv1beta1.AwsWebAclLabelMatchStatement{
			Key:   "aws:acl:block-high-risk",
			Scope: "LABEL",
		}

		result := convertLabelMatchStatement(labelMatch)
		assert.NotNil(t, result)
		assert.Equal(t, "aws:acl:block-high-risk", *result.Key)
		assert.Equal(t, wafv2types.LabelMatchScopeLabel, result.Scope)
	})

	t.Run("LabelMatch with NAMESPACE scope", func(t *testing.T) {
		labelMatch := &cloudresourcesv1beta1.AwsWebAclLabelMatchStatement{
			Key:   "aws:acl",
			Scope: "NAMESPACE",
		}

		result := convertLabelMatchStatement(labelMatch)
		assert.NotNil(t, result)
		assert.Equal(t, "aws:acl", *result.Key)
		assert.Equal(t, wafv2types.LabelMatchScopeNamespace, result.Scope)
	})
}

func TestConvertStatementWithLabelMatch(t *testing.T) {
	t.Run("Statement with LabelMatch", func(t *testing.T) {
		stmt := cloudresourcesv1beta1.AwsWebAclRuleStatement{
			LabelMatch: &cloudresourcesv1beta1.AwsWebAclLabelMatchStatement{
				Key:   "protection:advanced-logic",
				Scope: "LABEL",
			},
		}

		result, err := convertStatement(stmt)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.LabelMatchStatement)
		assert.Equal(t, "protection:advanced-logic", *result.LabelMatchStatement.Key)
		assert.Equal(t, wafv2types.LabelMatchScopeLabel, result.LabelMatchStatement.Scope)
	})
}

func TestConvertSizeConstraintStatement(t *testing.T) {
	t.Run("SizeConstraint with GT operator", func(t *testing.T) {
		sizeConstraint := &cloudresourcesv1beta1.AwsWebAclSizeConstraintStatement{
			ComparisonOperator: "GT",
			Size:               8192,
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				QueryString: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "NONE"},
			},
		}

		result, err := convertSizeConstraintStatement(sizeConstraint)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, wafv2types.ComparisonOperatorGt, result.ComparisonOperator)
		assert.Equal(t, int64(8192), result.Size)
		assert.NotNil(t, result.FieldToMatch.QueryString)
		assert.Len(t, result.TextTransformations, 1)
	})

	t.Run("SizeConstraint with LE operator on body", func(t *testing.T) {
		sizeConstraint := &cloudresourcesv1beta1.AwsWebAclSizeConstraintStatement{
			ComparisonOperator: "LE",
			Size:               1024,
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				Body: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "COMPRESS_WHITE_SPACE"},
			},
		}

		result, err := convertSizeConstraintStatement(sizeConstraint)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, wafv2types.ComparisonOperatorLe, result.ComparisonOperator)
		assert.Equal(t, int64(1024), result.Size)
		assert.NotNil(t, result.FieldToMatch.Body)
	})
}

func TestConvertStatementWithSizeConstraint(t *testing.T) {
	t.Run("Statement with SizeConstraint", func(t *testing.T) {
		stmt := cloudresourcesv1beta1.AwsWebAclRuleStatement{
			SizeConstraint: &cloudresourcesv1beta1.AwsWebAclSizeConstraintStatement{
				ComparisonOperator: "GT",
				Size:               10000,
				FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
					UriPath: true,
				},
				TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
					{Priority: 0, Type: "URL_DECODE"},
				},
			},
		}

		result, err := convertStatement(stmt)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.SizeConstraintStatement)
		assert.Equal(t, wafv2types.ComparisonOperatorGt, result.SizeConstraintStatement.ComparisonOperator)
		assert.Equal(t, int64(10000), result.SizeConstraintStatement.Size)
	})
}

func TestConvertSqliMatchStatement(t *testing.T) {
	t.Run("SqliMatch with LOW sensitivity", func(t *testing.T) {
		sqliMatch := &cloudresourcesv1beta1.AwsWebAclSqliMatchStatement{
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				QueryString: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "URL_DECODE"},
				{Priority: 1, Type: "HTML_ENTITY_DECODE"},
			},
			SensitivityLevel: "LOW",
		}

		result, err := convertSqliMatchStatement(sqliMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.FieldToMatch.QueryString)
		assert.Len(t, result.TextTransformations, 2)
		assert.Equal(t, wafv2types.SensitivityLevelLow, result.SensitivityLevel)
	})

	t.Run("SqliMatch with HIGH sensitivity", func(t *testing.T) {
		sqliMatch := &cloudresourcesv1beta1.AwsWebAclSqliMatchStatement{
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				Body: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "NONE"},
			},
			SensitivityLevel: "HIGH",
		}

		result, err := convertSqliMatchStatement(sqliMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.FieldToMatch.Body)
		assert.Equal(t, wafv2types.SensitivityLevelHigh, result.SensitivityLevel)
	})

	t.Run("SqliMatch with default sensitivity (empty string)", func(t *testing.T) {
		sqliMatch := &cloudresourcesv1beta1.AwsWebAclSqliMatchStatement{
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				UriPath: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "LOWERCASE"},
			},
			// SensitivityLevel not set
		}

		result, err := convertSqliMatchStatement(sqliMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, wafv2types.SensitivityLevelLow, result.SensitivityLevel)
	})
}

func TestConvertStatementWithSqliMatch(t *testing.T) {
	t.Run("Statement with SqliMatch", func(t *testing.T) {
		stmt := cloudresourcesv1beta1.AwsWebAclRuleStatement{
			SqliMatch: &cloudresourcesv1beta1.AwsWebAclSqliMatchStatement{
				FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
					QueryString: true,
				},
				TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
					{Priority: 0, Type: "URL_DECODE"},
				},
				SensitivityLevel: "HIGH",
			},
		}

		result, err := convertStatement(stmt)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.SqliMatchStatement)
		assert.Equal(t, wafv2types.SensitivityLevelHigh, result.SqliMatchStatement.SensitivityLevel)
	})
}

func TestConvertXssMatchStatement(t *testing.T) {
	t.Run("XssMatch with QueryString", func(t *testing.T) {
		xssMatch := &cloudresourcesv1beta1.AwsWebAclXssMatchStatement{
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				QueryString: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "URL_DECODE"},
				{Priority: 1, Type: "HTML_ENTITY_DECODE"},
			},
		}

		result, err := convertXssMatchStatement(xssMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.FieldToMatch.QueryString)
		assert.Len(t, result.TextTransformations, 2)
		assert.Equal(t, wafv2types.TextTransformationTypeUrlDecode, result.TextTransformations[0].Type)
		assert.Equal(t, wafv2types.TextTransformationTypeHtmlEntityDecode, result.TextTransformations[1].Type)
	})

	t.Run("XssMatch with Body", func(t *testing.T) {
		xssMatch := &cloudresourcesv1beta1.AwsWebAclXssMatchStatement{
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				Body: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "NONE"},
			},
		}

		result, err := convertXssMatchStatement(xssMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.FieldToMatch.Body)
		assert.Len(t, result.TextTransformations, 1)
	})

	t.Run("XssMatch with UriPath", func(t *testing.T) {
		xssMatch := &cloudresourcesv1beta1.AwsWebAclXssMatchStatement{
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				UriPath: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "LOWERCASE"},
				{Priority: 1, Type: "URL_DECODE"},
			},
		}

		result, err := convertXssMatchStatement(xssMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.FieldToMatch.UriPath)
		assert.Len(t, result.TextTransformations, 2)
	})
}

func TestConvertStatementWithXssMatch(t *testing.T) {
	t.Run("Statement with XssMatch", func(t *testing.T) {
		stmt := cloudresourcesv1beta1.AwsWebAclRuleStatement{
			XssMatch: &cloudresourcesv1beta1.AwsWebAclXssMatchStatement{
				FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
					QueryString: true,
				},
				TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
					{Priority: 0, Type: "URL_DECODE"},
					{Priority: 1, Type: "HTML_ENTITY_DECODE"},
				},
			},
		}

		result, err := convertStatement(stmt)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.XssMatchStatement)
		assert.NotNil(t, result.XssMatchStatement.FieldToMatch.QueryString)
	})
}

func TestConvertRegexMatchStatement(t *testing.T) {
	t.Run("RegexMatch with simple pattern", func(t *testing.T) {
		regexMatch := &cloudresourcesv1beta1.AwsWebAclRegexMatchStatement{
			RegexString: "^/api/v[0-9]+/.*$",
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				UriPath: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "LOWERCASE"},
			},
		}

		result, err := convertRegexMatchStatement(regexMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "^/api/v[0-9]+/.*$", *result.RegexString)
		assert.NotNil(t, result.FieldToMatch.UriPath)
		assert.Len(t, result.TextTransformations, 1)
	})

	t.Run("RegexMatch with email pattern", func(t *testing.T) {
		regexMatch := &cloudresourcesv1beta1.AwsWebAclRegexMatchStatement{
			RegexString: "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}",
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				Body: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "NONE"},
			},
		}

		result, err := convertRegexMatchStatement(regexMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, *result.RegexString, "@")
		assert.NotNil(t, result.FieldToMatch.Body)
	})

	t.Run("RegexMatch with multiple transformations", func(t *testing.T) {
		regexMatch := &cloudresourcesv1beta1.AwsWebAclRegexMatchStatement{
			RegexString: "(?i)(admin|root|superuser)",
			FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
				QueryString: true,
			},
			TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
				{Priority: 0, Type: "URL_DECODE"},
				{Priority: 1, Type: "LOWERCASE"},
			},
		}

		result, err := convertRegexMatchStatement(regexMatch)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.TextTransformations, 2)
	})
}

func TestConvertStatementWithRegexMatch(t *testing.T) {
	t.Run("Statement with RegexMatch", func(t *testing.T) {
		stmt := cloudresourcesv1beta1.AwsWebAclRuleStatement{
			RegexMatch: &cloudresourcesv1beta1.AwsWebAclRegexMatchStatement{
				RegexString: "^/admin/.*$",
				FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
					UriPath: true,
				},
				TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
					{Priority: 0, Type: "LOWERCASE"},
				},
			},
		}

		result, err := convertStatement(stmt)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.RegexMatchStatement)
		assert.Equal(t, "^/admin/.*$", *result.RegexMatchStatement.RegexString)
	})
}
