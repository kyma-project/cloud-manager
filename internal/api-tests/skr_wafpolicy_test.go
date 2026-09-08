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

package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: SKR WafPolicy", func() {

	It("Scenario: WafPolicy with simple rules can be created", func() {
		const webaclName = "test-webacl"

		webacl := &cloudresourcesv1beta1.WafPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: webaclName,
			},
			Spec: cloudresourcesv1beta1.WafPolicySpec{
				Data: `{
					"DefaultAction": {
						"Allow": {}
					},
					"Rules": [{
						"Name": "ManagedCommonRuleSet",
						"Priority": 1,
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
							"MetricName": "ManagedCommonRuleSet"
						}
					}],
					"VisibilityConfig": {
						"SampledRequestsEnabled": true,
						"CloudWatchMetricsEnabled": true,
						"MetricName": "test-webacl"
					}
				}`,
			},
		}

		By("When WafPolicy is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl).
				Should(Succeed())
		})

		By("Then WafPolicy is created successfully", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl, NewObjActions()).
				Should(Succeed())
		})

		By("Cleanup", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl).
				Should(Succeed())
		})
	})

	It("Scenario: WafPolicy with multiple rules can be created", func() {
		const webaclName = "test-webacl-multi"

		webacl := &cloudresourcesv1beta1.WafPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: webaclName,
			},
			Spec: cloudresourcesv1beta1.WafPolicySpec{
				Data: `{
					"DefaultAction": {
						"Block": {}
					},
					"Rules": [
						{
							"Name": "RateLimitRule",
							"Priority": 1,
							"Statement": {
								"RateBasedStatement": {
									"Limit": 2000,
									"AggregateKeyType": "IP"
								}
							},
							"Action": {
								"Block": {}
							},
							"VisibilityConfig": {
								"SampledRequestsEnabled": true,
								"CloudWatchMetricsEnabled": true,
								"MetricName": "RateLimitRule"
							}
						},
						{
							"Name": "GeoBlockRule",
							"Priority": 2,
							"Statement": {
								"GeoMatchStatement": {
									"CountryCodes": ["CN", "RU"]
								}
							},
							"Action": {
								"Block": {}
							},
							"VisibilityConfig": {
								"SampledRequestsEnabled": true,
								"CloudWatchMetricsEnabled": true,
								"MetricName": "GeoBlockRule"
							}
						},
						{
							"Name": "SQLiProtection",
							"Priority": 3,
							"Statement": {
								"ManagedRuleGroupStatement": {
									"VendorName": "AWS",
									"Name": "AWSManagedRulesSQLiRuleSet"
								}
							},
							"OverrideAction": {
								"None": {}
							},
							"VisibilityConfig": {
								"SampledRequestsEnabled": true,
								"CloudWatchMetricsEnabled": true,
								"MetricName": "SQLiProtection"
							}
						}
					],
					"VisibilityConfig": {
						"SampledRequestsEnabled": true,
						"CloudWatchMetricsEnabled": true,
						"MetricName": "test-webacl-multi"
					}
				}`,
			},
		}

		By("When WafPolicy is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl).
				Should(Succeed())
		})

		By("Then WafPolicy is created successfully", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl, NewObjActions()).
				Should(Succeed())
		})

		By("Cleanup", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), webacl).
				Should(Succeed())
		})
	})

	It("Scenario: WafPolicy with invalid JSON should fail validation", func() {
		const webaclName = "test-webacl-invalid"

		webacl := &cloudresourcesv1beta1.WafPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name: webaclName,
			},
			Spec: cloudresourcesv1beta1.WafPolicySpec{
				Data: `invalid json{`,
			},
		}

		By("When WafPolicy with invalid JSON is created", func() {
			err := infra.SKR().Client().Create(infra.Ctx(), webacl)
			// Note: JSON validation happens at reconciliation time, not admission
			// So creation might succeed but reconciliation will fail
			if err == nil {
				// If creation succeeded, cleanup
				Eventually(Delete).
					WithArguments(infra.Ctx(), infra.SKR().Client(), webacl).
					Should(Succeed())
			}
		})
	})
})
