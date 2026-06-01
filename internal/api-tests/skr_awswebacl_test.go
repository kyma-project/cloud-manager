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

	Context("Scenario: ManagedRuleGroup statement validation", func() {

		canCreateSkr(
			"AwsWebAcl can be created with single ManagedRuleGroup statement",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:           "managed-rule",
				Priority:       0,
				OverrideAction: cloudresourcesv1beta1.OverrideActionNone(), // Use OverrideAction for managed rules
				Statements: []cloudresourcesv1beta1.AwsWebAclStatement{
					{
						ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
							VendorName: "AWS",
							Name:       "AWSManagedRulesCommonRuleSet",
						},
					},
				},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with multiple managed rule groups",
			newTestAwsWebAclBuilder().WithRules([]cloudresourcesv1beta1.AwsWebAclRule{
				{
					Name:           "common-rules",
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
					Name:           "sql-injection-rules",
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
			}),
		)
	})

	Context("Scenario: Rule override action validation", func() {

		canCreateSkr(
			"AwsWebAcl can be created with None override action",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:           "none-rule",
				Priority:       0,
				OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
				Statements: []cloudresourcesv1beta1.AwsWebAclStatement{{
					ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
						VendorName: "AWS",
						Name:       "AWSManagedRulesCommonRuleSet",
					},
				}},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with Count override action",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:           "count-override-rule",
				Priority:       0,
				OverrideAction: cloudresourcesv1beta1.OverrideActionCount(),
				Statements: []cloudresourcesv1beta1.AwsWebAclStatement{{
					ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
						VendorName: "AWS",
						Name:       "AWSManagedRulesCommonRuleSet",
					},
				}},
			}),
		)

		canCreateSkr(
			"AwsWebAcl can be created with default override action (omitted)",
			newTestAwsWebAclBuilder().WithRule(cloudresourcesv1beta1.AwsWebAclRule{
				Name:     "default-override-rule",
				Priority: 0,
				// OverrideAction omitted - should default to None
				Statements: []cloudresourcesv1beta1.AwsWebAclStatement{{
					ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
						VendorName: "AWS",
						Name:       "AWSManagedRulesCommonRuleSet",
					},
				}},
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
					Name:           "new-rule",
					Priority:       0,
					OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
					Statements: []cloudresourcesv1beta1.AwsWebAclStatement{{
						ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
							VendorName: "AWS",
							Name:       "AWSManagedRulesCommonRuleSet",
						},
					}},
				})
			},
		)
	})
})
