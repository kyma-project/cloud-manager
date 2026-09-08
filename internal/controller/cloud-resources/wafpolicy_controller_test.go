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
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("WafPolicy Controller", func() {
	It("Scenario: SKR WafPolicy is created then deleted", func() {

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

		objName := "test-webacl"
		infra.ScopeProvider().Add(scopeprovider.MatchingObj(objName, scope))

		By("And Given scope is ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), scope,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed())
		})

		awsWebAcl := &cloudresourcesv1beta1.WafPolicy{}

		By("When WafPolicy is created", func() {
			awsWebAcl.Spec = cloudresourcesv1beta1.WafPolicySpec{
				Data: `{
					"DefaultAction": {
						"Allow": {}
					},
					"Rules": [
						{
							"Name": "KnownBadInputsRule",
							"Priority": 1,
							"Statement": {
								"ManagedRuleGroupStatement": {
									"VendorName": "AWS",
									"Name": "AWSManagedRulesKnownBadInputsRuleSet"
								}
							},
							"OverrideAction": {
								"None": {}
							},
							"VisibilityConfig": {
								"SampledRequestsEnabled": true,
								"CloudWatchMetricsEnabled": true,
								"MetricName": "KnownBadInputsRule"
							}
						},
						{
							"Name": "CommonRuleSet",
							"Priority": 2,
							"Statement": {
								"ManagedRuleGroupStatement": {
									"VendorName": "AWS",
									"Name": "AWSManagedRulesCommonRuleSet"
								}
							},
							"OverrideAction": {
								"None": {}
							},
							"VisibilityConfig": {
								"SampledRequestsEnabled": true,
								"CloudWatchMetricsEnabled": true,
								"MetricName": "CommonRuleSet"
							}
						}
					],
					"VisibilityConfig": {
						"SampledRequestsEnabled": true,
						"CloudWatchMetricsEnabled": true,
						"MetricName": "test-webacl"
					}
				}`,
			}
			awsWebAcl.ObjectMeta = metav1.ObjectMeta{
				Name: objName,
			}
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})

		By("Then WafPolicy gets Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsWebAcl,
					NewObjActions(),
					HavingState("Ready"),
				).
				Should(Succeed())
		})

		By("And Then WafPolicy has ARN in status", func() {
			Expect(awsWebAcl.Status.ProviderId).NotTo(BeEmpty())
		})

		// ========== Deletion ==========

		By("When WafPolicy is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})

		By("Then WafPolicy does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsWebAcl).
				Should(Succeed())
		})
	})
})
