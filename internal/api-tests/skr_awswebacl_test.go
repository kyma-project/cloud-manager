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
			WithDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultActionAllow).
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
			newTestAwsWebAclBuilder().WithDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultActionBlock),
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created with invalid default action",
			newTestAwsWebAclBuilder().WithDefaultAction("Invalid"),
			"spec.defaultAction",
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created without visibility config",
			&testAwsWebAclBuilder{
				AwsWebAclBuilder: cloudresourcesv1beta1.NewAwsWebAclBuilder().
					WithDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultActionAllow),
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
				Action:    cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{}, // Empty - no statement type
			}),
			"statement",
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created with multiple statement types",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "multiple-statements",
				Priority: 0,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{"10.0.0.0/8"},
					},
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"US"},
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionAllow,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{"10.0.0.0/8", "192.168.0.0/16"},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with single GeoMatch statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "geo-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					RateBased: &cloudresourcesv1beta1.AwsWebAclRateBasedStatement{
						Limit: 2000,
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with single ManagedRuleGroup statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "managed-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionCount,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "../",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							QueryString: true,
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "test",
						PositionalConstraint: "CONTAINS",
						FieldToMatch:         cloudresourcesv1beta1.AwsWebAclFieldToMatch{}, // Empty
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "test",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							UriPath:     true,
							QueryString: true, // Two fields set
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "test",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							UriPath:      true,
							SingleHeader: "user-agent", // Two fields set
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "/admin",
						PositionalConstraint: "STARTS_WITH",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							UriPath: true,
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "../",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							QueryString: true,
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "DELETE",
						PositionalConstraint: "EXACTLY",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							Method: true,
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "bot",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							SingleHeader: "user-agent",
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					ByteMatch: &cloudresourcesv1beta1.AwsWebAclByteMatchStatement{
						SearchString:         "<script>",
						PositionalConstraint: "CONTAINS",
						FieldToMatch: cloudresourcesv1beta1.AwsWebAclFieldToMatch{
							Body: true,
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionAllow,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{"10.0.0.0/8"},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with Block action",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "block-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{"192.0.2.0/24"},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with Count action",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "count-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionCount,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionCaptcha,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
						CountryCodes: []string{"CN"},
					},
				},
			}),
		)

		canNotCreateSkr(
			"AwsWebAcl cannot be created with invalid action",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "invalid-action",
				Priority: 0,
				Action:   "Invalid",
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{"10.0.0.0/8"},
					},
				},
			}),
			"action",
		)
	})

	Context("Scenario: IPSet statement validation", func() {

		canNotCreateSkr(
			"AwsWebAcl cannot be created with empty ipAddresses",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "empty-ips",
				Priority: 0,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{}, // Empty
					},
				},
			}),
			"ipAddresses",
		)

		canCreateSkr(
			"AwsWebAcl can be created with IPv4 addresses",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "ipv4-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{
							"10.0.0.0/8",
							"192.168.1.0/24",
							"203.0.113.42/32",
						},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with IPv6 addresses",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "ipv6-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{
							"2001:0db8::/32",
							"2001:0db8:85a3::8a2e:0370:7334/128",
						},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with mixed IPv4 and IPv6 addresses",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "mixed-ip-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionAllow,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{
							"10.0.0.0/8",
							"2001:0db8::/32",
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
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
					Action:   cloudresourcesv1beta1.AwsWebAclRuleActionAllow,
					Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
						IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
							IPAddresses: []string{"10.0.0.0/8"},
						},
					},
				},
				{
					Name:     "rule-2-geo",
					Priority: 1,
					Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
					Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
						GeoMatch: &cloudresourcesv1beta1.AwsWebAclGeoMatchStatement{
							CountryCodes: []string{"CN"},
						},
					},
				},
				{
					Name:     "rule-3-rate",
					Priority: 2,
					Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
					Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
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
			newTestAwsWebAclBuilder().WithDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultActionAllow),
			func(b Builder[*cloudresourcesv1beta1.AwsWebAcl]) {
				b.(*testAwsWebAclBuilder).WithDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultActionBlock)
			},
		)

		canChangeSkr(
			"AwsWebAcl defaultAction can be changed from Block to Allow",
			newTestAwsWebAclBuilder().WithDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultActionBlock),
			func(b Builder[*cloudresourcesv1beta1.AwsWebAcl]) {
				b.(*testAwsWebAclBuilder).WithDefaultAction(cloudresourcesv1beta1.AwsWebAclDefaultActionAllow)
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
					Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
					Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
						IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
							IPAddresses: []string{"10.0.0.0/8"},
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
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionBlock,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{"10.0.0.0/8"},
					},
				},
			}),
			func(b Builder[*cloudresourcesv1beta1.AwsWebAcl]) {
				webacl := b.Build()
				webacl.Spec.Rules[0].Statement.IPSet.IPAddresses = []string{
					"10.0.0.0/8",
					"192.168.0.0/16",
					"172.16.0.0/12",
				}
			},
		)

		canChangeSkr(
			"AwsWebAcl rule action can be changed",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "action-rule",
				Priority: 0,
				Action:   cloudresourcesv1beta1.AwsWebAclRuleActionCount,
				Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
					IPSet: &cloudresourcesv1beta1.AwsWebAclIPSetStatement{
						IPAddresses: []string{"10.0.0.0/8"},
					},
				},
			}),
			func(b Builder[*cloudresourcesv1beta1.AwsWebAcl]) {
				webacl := b.Build()
				webacl.Spec.Rules[0].Action = cloudresourcesv1beta1.AwsWebAclRuleActionBlock
			},
		)
	})
})
