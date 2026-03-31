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
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/wafv2/types"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AwsWebAcl Controller", Focus, func() {
	It("Scenario: SKR AwsWebAcl is created then deleted", func() {
		kymaName := infra.SkrKymaRef().Name

		awsAccountLocal := infra.AwsMock().NewAccount()
		defer awsAccountLocal.Delete()

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, awsAccountLocal.AccountId(), WithName(kymaName))).To(Succeed())
		})

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
				DefaultAction: cloudresourcesv1beta1.AwsWebAclDefaultActionAllow,
				Description:   "Web ACL for test application with AWS managed rule sets",
				VisibilityConfig: &cloudresourcesv1beta1.AwsWebAclVisibilityConfig{
					CloudWatchMetricsEnabled: true,
					MetricName:               "TestAppWAFMetrics",
					SampledRequestsEnabled:   true,
				},
				Rules: []cloudresourcesv1beta1.AwsWebAclRule{
					{
						Name:     "AWS-AWSManagedRulesBotControlRuleSet",
						Priority: 0,
						Action:   cloudresourcesv1beta1.AwsWebAclRuleActionCount,
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
						Name:     "AWS-AWSManagedRulesCommonRuleSet",
						Priority: 1,
						Action:   cloudresourcesv1beta1.AwsWebAclRuleActionCount,
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
					WithName("waf-for-test-app"),
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

		By("And Then WebACL exists in AWS mock with correct configuration", func() {
			// Extract ID from ARN: arn:aws:wafv2:region:account:scope/webacl/name/id
			arnParts := strings.Split(awsWebAcl.Status.Arn, "/")
			Expect(len(arnParts)).To(BeNumerically(">=", 4), "expected ARN to have at least 4 parts")
			id := arnParts[len(arnParts)-1]

			awsWebACL, _, err := awsMockLocal.GetWebACL(infra.Ctx(), awsWebAcl.Name, id, types.ScopeRegional)
			Expect(err).NotTo(HaveOccurred(), "expected WebACL to exist in mock")
			Expect(awsWebACL).NotTo(BeNil())
			Expect(*awsWebACL.Name).To(Equal(awsWebAcl.Name))
			Expect(*awsWebACL.Description).To(Equal(awsWebAcl.Spec.Description))
			Expect(awsWebACL.Rules).To(HaveLen(2), "expected 2 rules in mock")
			Expect(*awsWebACL.Rules[0].Name).To(Equal("AWS-AWSManagedRulesBotControlRuleSet"))
			Expect(*awsWebACL.Rules[1].Name).To(Equal("AWS-AWSManagedRulesCommonRuleSet"))
			Expect(awsWebACL.Capacity).To(Equal(int64(100)), "expected mock capacity to be 100")
		})

		arn := ""

		By("When AwsWebAcl is deleted", func() {
			arn = awsWebAcl.Status.Arn // Save ARN before deletion clears it
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})

		By("Then AwsWebAcl is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})

		By("And Then WebACL is deleted from AWS mock", func() {
			// Extract ID from saved ARN
			arnParts := strings.Split(arn, "/")
			id := arnParts[len(arnParts)-1]

			_, _, err := awsMockLocal.GetWebACL(infra.Ctx(), "waf-for-test-app", id, types.ScopeRegional)
			Expect(err).To(HaveOccurred(), "expected WebACL to be deleted from mock")
			Expect(err.Error()).To(ContainSubstring("WAFNonexistentItemException"), "expected not found error")
		})
	})
})
