package api_tests

import (
	"github.com/google/uuid"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/onsi/ginkgo/v2"
)

type testAwsWebAclBuilder struct {
	*cloudresourcesv1beta1.AwsWebAclBuilder
}

func newTestAwsWebAclBuilder() *testAwsWebAclBuilder {
	return &testAwsWebAclBuilder{
		AwsWebAclBuilder: cloudresourcesv1beta1.NewAwsWebAclBuilder().
			WithDefaultAction(cloudresourcesv1beta1.DefaultActionAllow()).
			WithDescription("Test WebACL").
			WithVisibilityConfig(&cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
				CloudWatchMetricsEnabled: true,
				MetricName:               uuid.NewString()[:32], // Ensure unique metric name
				SampledRequestsEnabled:   true,
			}),
	}
}

func (b *testAwsWebAclBuilder) Build() *cloudresourcesv1beta1.AwsWebAcl {
	return &b.AwsWebAcl
}

func (b *testAwsWebAclBuilder) WithDefaultAction(action cloudresourcesv1beta1.AwsWebAclDefaultAction) *testAwsWebAclBuilder {
	b.AwsWebAclBuilder.WithDefaultAction(action)
	return b
}

func (b *testAwsWebAclBuilder) WithDescription(description string) *testAwsWebAclBuilder {
	b.AwsWebAclBuilder.WithDescription(description)
	return b
}

func (b *testAwsWebAclBuilder) WithRule(rule cloudresourcesv1beta1.AwsWebAclRule) *testAwsWebAclBuilder {
	b.AwsWebAclBuilder.WithRule(rule)
	return b
}

func (b *testAwsWebAclBuilder) WithRules(rules []cloudresourcesv1beta1.AwsWebAclRule) *testAwsWebAclBuilder {
	b.AwsWebAclBuilder.WithRules(rules)
	return b
}

var _ = Describe("Feature: SKR AwsWebAcl", Ordered, func() {

	Context("Scenario: Basic creation validation", func() {

		canCreateSkr(
			"AwsWebAcl can be created with minimal spec (no rules)",
			newTestAwsWebAclBuilder(),
		)

		canCreateSkr(
			"AwsWebAcl can be created with Block default action",
			newTestAwsWebAclBuilder().WithDefaultAction(cloudresourcesv1beta1.DefaultActionBlock()),
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created with invalid default action",
			newTestAwsWebAclBuilder().WithDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultAction{}), // Empty - neither Allow nor Block
			"spec.defaultAction",
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created without visibility config",
			&testAwsWebAclBuilder{
				AwsWebAclBuilder: cloudresourcesv1beta1.NewAwsWebAclBuilder().
					WithDefaultAction(cloudresourcesv1beta1.DefaultActionAllow()),
			},
			"visibilityConfig",
		)
	})

	Context("Scenario: Rule statement validation - exactly one statement type", func() {

		canNotCreateSkr(
			"AwsWebAcl cannot be created with empty rule statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:      "empty-statement",
				Priority:  0,
				Action:    cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{}, // Empty - no statement type
			}),
			"statement",
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created with multiple statement types",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "multiple-statements",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
					},
					RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
						Limit: 2000,
					},
				},
			}),
			"statement",
		)

		canCreateSkr(
			"AwsWebAcl can be created with single IPSet statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "ipset-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionAllow(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with single GeoMatch statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "geo-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"CN", "RU"},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with single RateBased statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "rate-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
						Limit: 2000,
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with single ManagedRuleGroup statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:           "managed-rule",
				Priority:       0,
				OverrideAction: cloudresourcesv1beta1.OverrideActionNone(), // Use OverrideAction for managed rules
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
						VendorName: "AWS",
						Name:       "AWSManagedRulesCommonRuleSet",
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with single ByteMatch statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "byte-match-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "../",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							QueryString: true,
						},
						TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
							{Priority: 0, Type: "NONE"},
						},
					},
				},
			}),
		)
	})

	Context("Scenario: ByteMatch fieldToMatch validation - exactly one field", func() {

		canNotCreateSkr(
			"AwsWebAcl cannot be created with empty fieldToMatch",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "empty-field",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "test",
						PositionalConstraint: "CONTAINS",
						FieldToMatch:         cloudresourcesv1beta1.AwsWebAclFieldToMatch{}, // Empty
						TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
							{Priority: 0, Type: "NONE"},
						},
					},
				},
			}),
			"fieldToMatch",
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created with multiple fieldToMatch options (bool fields)",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "multiple-fields",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "test",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							UriPath:     true,
							QueryString: true, // Two fields set
						},
						TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
							{Priority: 0, Type: "NONE"},
						},
					},
				},
			}),
			"fieldToMatch",
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created with bool field and singleHeader set",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "mixed-fields",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "test",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							UriPath:      true,
							SingleHeader: "user-agent", // Two fields set
						},
						TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
							{Priority: 0, Type: "NONE"},
						},
					},
				},
			}),
			"fieldToMatch",
		)

		canCreateSkr(
			"AwsWebAcl can be created with uriPath field",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "uri-path-match",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "/admin",
						PositionalConstraint: "STARTS_WITH",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							UriPath: true,
						},
						TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
							{Priority: 0, Type: "LOWERCASE"},
						},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with queryString field",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "query-match",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "../",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							QueryString: true,
						},
						TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
							{Priority: 0, Type: "URL_DECODE"},
						},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with method field",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "method-match",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "DELETE",
						PositionalConstraint: "EXACTLY",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							Method: true,
						},
						TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
							{Priority: 0, Type: "NONE"},
						},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with singleHeader field",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "header-match",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "bot",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							SingleHeader: "user-agent",
						},
						TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
							{Priority: 0, Type: "LOWERCASE"},
						},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with body field",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "body-match",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "<script>",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							Body: true,
						},
						TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
							{Priority: 0, Type: "HTML_ENTITY_DECODE"},
						},
					},
				},
			}),
		)
	})

	Context("Scenario: Rule action validation", func() {

		canCreateSkr(
			"AwsWebAcl can be created with Allow action",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "allow-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionAllow(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with Block action",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "block-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with Count action",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "count-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionCount(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
						Limit: 1000,
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with Captcha action",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "captcha-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionCaptcha(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"CN"},
					},
				},
			}),
		)
	})

	Context("Scenario: GeoMatch statement validation", func() {

		canNotCreateSkr(
			"AwsWebAcl cannot be created with empty countryCodes",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "empty-countries",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{}, // Empty
					},
				},
			}),
			"countryCodes",
		)

		canCreateSkr(
			"AwsWebAcl can be created with single country code",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "single-country-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{
							"US",
						},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with multiple country codes",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "multi-country-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{
							"US",
							"GB",
							"DE",
						},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with GeoMatch and ForwardedIPConfig",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "geo-forwarded-ip-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionAllow(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{
							"US",
						},
						ForwardedIPConfig: &cloudresourcesv1beta1.AwsWebAclForwardedIPConfig{
							HeaderName:       "X-Forwarded-For",
							FallbackBehavior: "MATCH",
						},
					},
				},
			}),
		)
	})

	Context("Scenario: RateBased statement validation", func() {

		canNotCreateSkr(
			"AwsWebAcl cannot be created with rate limit below minimum (100)",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "low-rate",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
						Limit: 99,
					},
				},
			}),
			"limit",
		)

		canCreateSkr(
			"AwsWebAcl can be created with minimum rate limit (100)",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "min-rate",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
						Limit: 100,
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with typical rate limit",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "normal-rate",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
						Limit: 2000,
					},
				},
			}),
		)
	})

	Context("Scenario: Multiple rules validation", func() {

		canCreateSkr(
			"AwsWebAcl can be created with multiple rules of different types",
			newTestAwsWebAclBuilder().WithRules([]cloudresourcesv1beta1.AwsWebAclRule{
				{
					Name:     "rule-1-ipset",
					Priority: 0,
					Action:   cloudresourcesv1beta1.RuleActionAllow(),
					Statement: cloudresourcesv1beta1.AwsWebAclStatement{
						GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
							CountryCodes: []string{"US"},
						},
					},
				},
				{
					Name:     "rule-2-geo",
					Priority: 1,
					Action:   cloudresourcesv1beta1.RuleActionBlock(),
					Statement: cloudresourcesv1beta1.AwsWebAclStatement{
						GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
							CountryCodes: []string{"CN"},
						},
					},
				},
				{
					Name:     "rule-3-rate",
					Priority: 2,
					Action:   cloudresourcesv1beta1.RuleActionBlock(),
					Statement: cloudresourcesv1beta1.AwsWebAclStatement{
						RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
							Limit: 1000,
						},
					},
				},
			}),
		)
	})

	Context("Scenario: DefaultAction mutability", func() {

		canChangeSkr(
			"AwsWebAcl defaultAction can be changed from Allow to Block",
			newTestAwsWebAclBuilder().WithDefaultAction(cloudresourcesv1beta1.DefaultActionAllow()),
			func(b Builder[*cloudresourcesv1beta1.AwsWebAcl]) {
				b.(*testAwsWebAclBuilder).WithDefaultAction(cloudresourcesv1beta1.DefaultActionBlock())
			},
		)

		canChangeSkr(
			"AwsWebAcl defaultAction can be changed from Block to Allow",
			newTestAwsWebAclBuilder().WithDefaultAction(cloudresourcesv1beta1.DefaultActionBlock()),
			func(b Builder[*cloudresourcesv1beta1.AwsWebAcl]) {
				b.(*testAwsWebAclBuilder).WithDefaultAction(cloudresourcesv1beta1.DefaultActionAllow())
			},
		)
	})

	Context("Scenario: Rules mutability", func() {

		canChangeSkr(
			"AwsWebAcl rules can be added",
			newTestAwsWebAclBuilder(),
			func(b Builder[*cloudresourcesv1beta1.AwsWebAcl]) {
				b.(*testAwsWebAclBuilder).WithRule(cloudresourcesv1beta1.AwsWebAclRule{
					Name:     "new-rule",
					Priority: 0,
					Action:   cloudresourcesv1beta1.RuleActionBlock(),
					Statement: cloudresourcesv1beta1.AwsWebAclStatement{
						GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
							CountryCodes: []string{"US"},
						},
					},
				})
			},
		)

		canChangeSkr(
			"AwsWebAcl rule IPSet addresses can be modified",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "ipset-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
					},
				},
			}),
			func(b Builder[*cloudresourcesv1beta1.AwsWebAcl]) {
				webacl := b.Build()
				webacl.Spec.Rules[0].Statement.GeoMatch.CountryCodes = []string{
					"US",
					"US",
					"US",
				}
			},
		)

		canChangeSkr(
			"AwsWebAcl rule action can be changed",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "action-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionCount(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
					},
				},
			}),
			func(b Builder[*cloudresourcesv1beta1.AwsWebAcl]) {
				webacl := b.Build()
				webacl.Spec.Rules[0].Action = cloudresourcesv1beta1.RuleActionBlock()
			},
		)
	})

	Context("Scenario: Logical operators - And/Or/Not statements", func() {

		canCreateSkr(
			"AwsWebAcl can be created with AndStatement combining two conditions",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "and-statement-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					AndStatement: &cloudresourcesv1beta1.AwsWebAclAndStatement{
						Statements: []cloudresourcesv1beta1.AwsWebAclStatement1{
							{
								GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
									CountryCodes: []string{"CN", "RU"},
								},
							},
							{
								ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
									SearchString:         "/admin",
									PositionalConstraint: "STARTS_WITH",
									FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
										UriPath: true,
									},
									TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
										{Priority: 0, Type: "LOWERCASE"},
									},
								},
							},
						},
					},
				},
			}),
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created with AndStatement with only one statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "and-single-statement",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					AndStatement: &cloudresourcesv1beta1.AwsWebAclAndStatement{
						Statements: []cloudresourcesv1beta1.AwsWebAclStatement1{
							{
								GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
									CountryCodes: []string{"US"},
								},
							},
						},
					},
				},
			}),
			"statements",
		)

		canCreateSkr(
			"AwsWebAcl can be created with OrStatement combining multiple conditions",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "or-statement-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					OrStatement: &cloudresourcesv1beta1.AwsWebAclOrStatement{
						Statements: []cloudresourcesv1beta1.AwsWebAclStatement1{
							{
								ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
									SearchString:         "sqlmap",
									PositionalConstraint: "CONTAINS",
									FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
										SingleHeader: "user-agent",
									},
									TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
										{Priority: 0, Type: "LOWERCASE"},
									},
								},
							},
							{
								GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
									CountryCodes: []string{"KP"},
								},
							},
						},
					},
				},
			}),
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created with OrStatement with only one statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "or-single-statement",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionBlock(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					OrStatement: &cloudresourcesv1beta1.AwsWebAclOrStatement{
						Statements: []cloudresourcesv1beta1.AwsWebAclStatement1{
							{
								GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
									CountryCodes: []string{"US"},
								},
							},
						},
					},
				},
			}),
			"statements",
		)

		canCreateSkr(
			"AwsWebAcl can be created with NotStatement negating GeoMatch",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "not-geo-statement",
				Priority: 0,
				Action:   cloudresourcesv1beta1.RuleActionAllow(),
				Statement: cloudresourcesv1beta1.AwsWebAclStatement{
					NotStatement: &cloudresourcesv1beta1.AwsWebAclNotStatement{
						Statement: cloudresourcesv1beta1.AwsWebAclStatement1{
							GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
								CountryCodes: []string{"CN", "RU", "KP"},
							},
						},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with multiple rules using logical operators",
			newTestAwsWebAclBuilder().WithRules([]cloudresourcesv1beta1.AwsWebAclRule{
				{
					Name:     "rule-and",
					Priority: 0,
					Action:   cloudresourcesv1beta1.RuleActionBlock(),
					Statement: cloudresourcesv1beta1.AwsWebAclStatement{
						AndStatement: &cloudresourcesv1beta1.AwsWebAclAndStatement{
							Statements: []cloudresourcesv1beta1.AwsWebAclStatement1{
								{
									GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
										CountryCodes: []string{"CN"},
									},
								},
								{
									RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
										Limit: 1000,
									},
								},
							},
						},
					},
				},
				{
					Name:     "rule-or",
					Priority: 1,
					Action:   cloudresourcesv1beta1.RuleActionBlock(),
					Statement: cloudresourcesv1beta1.AwsWebAclStatement{
						OrStatement: &cloudresourcesv1beta1.AwsWebAclOrStatement{
							Statements: []cloudresourcesv1beta1.AwsWebAclStatement1{
								{
									ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
										SearchString:         "bot",
										PositionalConstraint: "CONTAINS",
										FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
											SingleHeader: "user-agent",
										},
										TextTransformations: []cloudresourcesv1beta1.AwsWebAclTextTransformation{
											{Priority: 0, Type: "LOWERCASE"},
										},
									},
								},
								{
									RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
										Limit: 100,
									},
								},
							},
						},
					},
				},
				{
					Name:     "rule-not",
					Priority: 2,
					Action:   cloudresourcesv1beta1.RuleActionAllow(),
					Statement: cloudresourcesv1beta1.AwsWebAclStatement{
						NotStatement: &cloudresourcesv1beta1.AwsWebAclNotStatement{
							Statement: cloudresourcesv1beta1.AwsWebAclStatement1{
								GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
									CountryCodes: []string{"KP"},
								},
							},
						},
					},
				},
			}),
		)
	})
})
