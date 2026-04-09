/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cloudresources

import (
	"fmt"
	"strings"

	wafv2types "github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	awsutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/aws/util"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Test helper ObjActions
type setAwsWebAclDefaultAction struct {
	action cloudresourcesv1beta1.AwsWebAclDefaultAction
}

func (a *setAwsWebAclDefaultAction) Apply(obj client.Object) {
	obj.(*cloudresourcesv1beta1.AwsWebAcl).Spec.DefaultAction = a.action
}

func SetAwsWebAclDefaultAction(action cloudresourcesv1beta1.AwsWebAclDefaultAction) ObjAction {
	return &setAwsWebAclDefaultAction{action: action}
}

type addAwsWebAclRule struct {
	rule cloudresourcesv1beta1.AwsWebAclRule
}

func (a *addAwsWebAclRule) Apply(obj client.Object) {
	acl := obj.(*cloudresourcesv1beta1.AwsWebAcl)
	acl.Spec.Rules = append(acl.Spec.Rules, a.rule)
}

func AddAwsWebAclRule(rule cloudresourcesv1beta1.AwsWebAclRule) ObjAction {
	return &addAwsWebAclRule{rule: rule}
}

var _ = Describe("AwsWebAcl Controller", func() {
	It("Scenario: SKR AwsWebAcl is created then deleted", func() {

		awsAccountLocal := infra.AwsMock().NewAccount()
		defer awsAccountLocal.Delete()

		scope := &cloudcontrolv1beta1.Scope{}

		scopeName := "e08b6fe8-9628-4601-8351-7d443a078606"

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(scopeName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, awsAccountLocal.AccountId(), WithName(scopeName))).To(Succeed())
		})

		Expect(scope.Namespace).To(Equal(infra.KCP().Namespace()))
		Expect(scope.Name).To(Equal(scopeName))

		objName := "nam"
		infra.ScopeProvider().Add(scopeprovider.MatchingObj(objName, scope))

		awsMockLocal := awsAccountLocal.Region(scope.Spec.Region)

		By("And Given scope is ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		awsWebAcl := &cloudresourcesv1beta1.AwsWebAcl{}

		By("When AwsWebAcl is created", func() {
			// Create with comprehensive spec similar to the sample CRD
			awsWebAcl.Spec = cloudresourcesv1beta1.AwsWebAclSpec{
				DefaultAction: cloudresourcesv1beta1.DefaultActionAllow(),
				Description:   "Web ACL for test application with AWS managed rule sets",
				VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
					CloudWatchMetricsEnabled: true,
					MetricName:               "TestAppWAFMetrics",
					SampledRequestsEnabled:   true,
				},
				Rules: []cloudresourcesv1beta1.AwsWebAclRule{
					{
						Name:           "AWS-AWSManagedRulesBotControlRuleSet",
						Priority:       0,
						OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
						Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
							ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
								VendorName: "AWS",
								Name:       "AWSManagedRulesBotControlRuleSet",
							},
						},
						VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
							CloudWatchMetricsEnabled: true,
							MetricName:               "AWS-AWSManagedRulesBotControlRuleSet",
							SampledRequestsEnabled:   true,
						},
					},
					{
						Name:           "AWS-AWSManagedRulesCommonRuleSet",
						Priority:       1,
						OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
						Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
							ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
								VendorName: "AWS",
								Name:       "AWSManagedRulesCommonRuleSet",
							},
						},
						VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
							CloudWatchMetricsEnabled: true,
							MetricName:               "AWS-AWSManagedRulesCommonRuleSet",
							SampledRequestsEnabled:   true,
						},
					},
				},
			}

			Eventually(CreateAwsWebAcl).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsWebAcl,
					WithName(objName),
				).Should(Succeed())
		})

		By("Then AwsWebAcl is loaded and has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsWebAcl,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("And Then AwsWebAcl status has ARN populated", func() {
			Expect(awsWebAcl.Status.Arn).NotTo(BeEmpty(), "expected Status.Arn to be set")
			Expect(awsWebAcl.Status.Arn).To(ContainSubstring("arn:aws:wafv2:"), "expected ARN to have correct format")
			Expect(awsWebAcl.Status.Arn).To(ContainSubstring(":regional/webacl/"), "expected ARN to reference WebACL")
		})

		By("And Then AwsWebAcl status has Capacity populated", func() {
			Expect(awsWebAcl.Status.Capacity).To(BeNumerically(">", 0), "expected Capacity to be greater than 0")
		})

		By("And Then WebACL exists in AWS with correct configuration", func() {
			// Extract ID from ARN: arn:aws:wafv2:region:account:scope/webacl/name/id
			arnParts := strings.Split(awsWebAcl.Status.Arn, "/")
			Expect(len(arnParts)).To(BeNumerically(">=", 4), "expected ARN to have at least 4 parts")
			id := arnParts[len(arnParts)-1]

			awsWebACL, _, err := awsMockLocal.GetWebACL(infra.Ctx(), awsWebAcl.Name, id, wafv2types.ScopeRegional)
			Expect(err).NotTo(HaveOccurred(), "expected WebACL to exist")
			Expect(awsWebACL).NotTo(BeNil())
			Expect(*awsWebACL.Name).To(Equal(awsWebAcl.Name))
			Expect(*awsWebACL.Description).To(Equal(awsWebAcl.Spec.Description))
			Expect(awsWebACL.Rules).To(HaveLen(2), "expected 2 rules")
			Expect(*awsWebACL.Rules[0].Name).To(Equal("AWS-AWSManagedRulesBotControlRuleSet"))
			Expect(*awsWebACL.Rules[1].Name).To(Equal("AWS-AWSManagedRulesCommonRuleSet"))
			Expect(awsWebACL.Capacity).To(Equal(int64(100)), "expected capacity to be 100")
		})

		id := awsutil.ParseArnResourceId(awsWebAcl.Status.Arn)

		By("When AwsWebAcl spec is updated to change DefaultAction", func() {
			Eventually(Update).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl,
					SetAwsWebAclDefaultAction(cloudresourcesv1beta1.DefaultActionBlock()),
				).Should(Succeed())

		})

		By("And Then WebACL in AWS has updated DefaultAction", func() {
			Eventually(func() error {
				awsWebACL, _, err := awsMockLocal.GetWebACL(infra.Ctx(), awsWebAcl.Name, id, wafv2types.ScopeRegional)
				if err != nil {
					return err
				}
				if awsWebACL.DefaultAction == nil {
					return fmt.Errorf("DefaultAction is nil")
				}
				if awsWebACL.DefaultAction.Block == nil {
					return fmt.Errorf("expected Block action, got: %+v", awsWebACL.DefaultAction)
				}
				if awsWebACL.DefaultAction.Allow != nil {
					return fmt.Errorf("expected Allow action to be nil")
				}
				return nil
			}).Should(Succeed())
		})

		By("Then AwsWebAcl returns to Ready state after update", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsWebAcl,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("When AwsWebAcl spec is updated to add a new rule", func() {
			Eventually(Update).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl,
					AddAwsWebAclRule(cloudresourcesv1beta1.AwsWebAclRule{
						Name:           "AWS-AWSManagedRulesKnownBadInputsRuleSet",
						Priority:       2,
						OverrideAction: cloudresourcesv1beta1.OverrideActionNone(),
						Statement: cloudresourcesv1beta1.AwsWebAclRuleStatement{
							ManagedRuleGroup: &cloudresourcesv1beta1.AwsWebAclManagedRuleGroupStatement{
								VendorName: "AWS",
								Name:       "AWSManagedRulesKnownBadInputsRuleSet",
							},
						},
						VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
							CloudWatchMetricsEnabled: true,
							MetricName:               "AWS-AWSManagedRulesKnownBadInputsRuleSet",
							SampledRequestsEnabled:   true,
						},
					}),
				).Should(Succeed())

		})

		By("And Then WebACL in AWS has 3 rules", func() {
			Eventually(func() error {
				awsWebACL, _, err := awsMockLocal.GetWebACL(infra.Ctx(), awsWebAcl.Name, id, wafv2types.ScopeRegional)
				if err != nil {
					return err
				}
				if len(awsWebACL.Rules) != 3 {
					return fmt.Errorf("expected 3 rules, got %d", len(awsWebACL.Rules))
				}
				if *awsWebACL.Rules[2].Name != "AWS-AWSManagedRulesKnownBadInputsRuleSet" {
					return fmt.Errorf("expected third rule to be AWS-AWSManagedRulesKnownBadInputsRuleSet, got %s", *awsWebACL.Rules[2].Name)
				}
				return nil
			}).Should(Succeed())
		})

		By("Then AwsWebAcl returns to Ready state after adding rule", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsWebAcl,
					NewObjActions(),
					HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("When AwsWebAcl is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})

		By("Then AwsWebAcl is does not exists", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})

		By("And Then WebACL is deleted from AWS", func() {
			_, _, err := awsMockLocal.GetWebACL(infra.Ctx(), objName, id, wafv2types.ScopeRegional)
			Expect(err).To(HaveOccurred(), "expected WebACL to be deleted")
			Expect(err.Error()).To(ContainSubstring("WAFNonexistentItemException"), "expected not found error")

		})
	})
})
