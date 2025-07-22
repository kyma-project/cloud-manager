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

package cloudcontrol

import (
	"fmt"
	"strings"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpiprange "github.com/kyma-project/cloud-manager/pkg/kcp/iprange"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
)

var _ = (client.Client)(nil)

var _ = Describe("Feature: KCP Subscription", func() {

	commonInit := func(provider cloudcontrolv1beta1.ProviderType, subscriptionName string, secretData map[string][]byte) (*corev1.Secret, *gardenertypes.SecretBinding, *cloudcontrolv1beta1.Subscription) {
		secret := &corev1.Secret{}
		secret.Name = subscriptionName
		secret.Namespace = DefaultGardenNamespace
		secret.Data = secretData

		secretBinding := &gardenertypes.SecretBinding{}
		secretBinding.Name = subscriptionName
		secretBinding.Namespace = DefaultGardenNamespace
		secretBinding.Provider = &gardenertypes.SecretBindingProvider{Type: string(provider)}
		secretBinding.SecretRef.Name = subscriptionName
		secretBinding.SecretRef.Namespace = DefaultGardenNamespace

		subscription := cloudcontrolv1beta1.NewSubscriptionBuilder().
			WithName(subscriptionName).
			WithNamespace(DefaultKcpNamespace).
			WithSecretBindingName(secretBinding.Name).
			Build()

		By(fmt.Sprintf("Given Garden Secret with %s credentials exists", strings.ToUpper(string(provider))), func() {
			Expect(CreateObj(infra.Ctx(), infra.Garden().Client(), secret)).To(Succeed())
		})

		By("Given Garden SecretBinding exists", func() {
			Expect(CreateObj(infra.Ctx(), infra.Garden().Client(), secretBinding)).To(Succeed())
		})

		By("Given KCP secret gardener-credentials exist", func() {
			Expect(CreateGardenerCredentials(infra.Ctx(), infra)).To(Succeed())
		})

		return secret, secretBinding, subscription
	}

	It("Scenario: KCP Subscription AWS is created and deleted", func() {

		const (
			provider         = cloudcontrolv1beta1.ProviderAws
			subscriptionName = "1c3a1ec7-3558-467b-a127-25da69fc1887"
			awsAccountId     = "66beaa3c-69b9-4617-a7c4-4a10dca9ad48"
		)

		infra.AwsMock().SetAccount(awsAccountId)

		secret, secretBinding, subscription := commonInit(provider, subscriptionName, map[string][]byte{
			"accessKeyID":     []byte("some-key-id"),
			"secretAccessKey": []byte("some-secret-access-key"),
		})

		By("When KCP Subscription is created", func() {
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), subscription)).To(Succeed())
		})

		By("Then KCP Subscription is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		By("And Then KCP Subscription status.provider is set", func() {
			Expect(subscription.Status.Provider).To(Equal(provider))
		})

		By("And Then KCP Subscription status.subscriptionInfo.aws is set", func() {
			Expect(subscription.Status.SubscriptionInfo).NotTo(BeNil())
			Expect(subscription.Status.SubscriptionInfo.Aws).NotTo(BeNil())
			Expect(subscription.Status.SubscriptionInfo.Aws.Account).To(Equal(awsAccountId))
		})

		By("And Then KCP Subscription has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(subscription, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		By("When KCP Subscription is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), subscription)).To(Succeed())
		})

		By("Then KCP Subscription does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription).
				Should(Succeed())
		})

		By("// cleanup: Delete Secret and SecretBinding", func() {
			Expect(Delete(infra.Ctx(), infra.Garden().Client(), secret)).To(Succeed())
			Expect(Delete(infra.Ctx(), infra.Garden().Client(), secretBinding)).To(Succeed())
		})
	})

	It("Scenario: KCP Subscription AWS deletion is blocked when used", func() {

		const (
			provider         = cloudcontrolv1beta1.ProviderAws
			subscriptionName = "f5591f18-a6d0-4864-b08c-5a874023be2e"
			awsAccountId     = "227b765f-f4bb-480f-b205-7aef384fd712"
		)

		infra.AwsMock().SetAccount(awsAccountId)

		secret, secretBinding, subscription := commonInit(provider, subscriptionName, map[string][]byte{
			"accessKeyID":     []byte("some-key-id"),
			"secretAccessKey": []byte("some-secret-access-key"),
		})

		By("When KCP Subscription is created", func() {
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), subscription)).To(Succeed())
		})

		By("Then KCP Subscription is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		var resources []client.Object

		By("When KCP resources labeled to use subscription are created", func() {
			// Random kinds used, kind is not important, but only the label `cloudcontrolv1beta1.SubscriptionLabel: subscriptionName`
			var x client.Object
			kcpscope.Ignore.AddName(subscriptionName)
			x = &cloudcontrolv1beta1.Scope{
				ObjectMeta: metav1.ObjectMeta{
					Name:      subscriptionName,
					Namespace: DefaultKcpNamespace,
					Labels: map[string]string{
						cloudcontrolv1beta1.SubscriptionLabel: subscriptionName,
					},
				},
				Spec: cloudcontrolv1beta1.ScopeSpec{
					KymaName:  subscriptionName,
					ShootName: subscriptionName,
					Region:    "eu-central-1",
					Provider:  cloudcontrolv1beta1.ProviderAws,
					Scope: cloudcontrolv1beta1.ScopeInfo{
						Aws: &cloudcontrolv1beta1.AwsScope{
							VpcNetwork: subscriptionName,
							AccountId:  awsAccountId,
							Network: cloudcontrolv1beta1.AwsNetwork{
								VPC:   cloudcontrolv1beta1.AwsVPC{},
								Zones: []cloudcontrolv1beta1.AwsZone{},
							},
						},
					},
				},
			}
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), x)).To(Succeed())
			resources = append(resources, x)

			kcpiprange.Ignore.AddName(subscriptionName)
			x = &cloudcontrolv1beta1.IpRange{
				ObjectMeta: metav1.ObjectMeta{
					Name:      subscriptionName,
					Namespace: DefaultKcpNamespace,
					Labels: map[string]string{
						cloudcontrolv1beta1.SubscriptionLabel: subscriptionName,
					},
				},
				Spec: cloudcontrolv1beta1.IpRangeSpec{
					Scope:     cloudcontrolv1beta1.ScopeRef{Name: subscriptionName},
					RemoteRef: cloudcontrolv1beta1.RemoteRef{Name: "a", Namespace: "b"},
				},
			}
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), x)).To(Succeed())
			resources = append(resources, x)
		})

		By("When KCP Subscription is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), subscription)).To(Succeed())
		})

		By("Then KCP Subscription still exists with Warning/DeleteWhileUsed condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions(), HavingConditionReasonTrue(cloudcontrolv1beta1.ConditionTypeWarning, cloudcontrolv1beta1.ReasonDeleteWhileUsed)).
				Should(Succeed())
			cond := meta.FindStatusCondition(subscription.Status.Conditions, cloudcontrolv1beta1.ConditionTypeWarning)
			// Used by: cloud-control.kyma-project.io/v1beta1/IpRange: f5591f18-a6d0-4864-b08c-5a874023be2., cloud-control.kyma-project.io/v1beta1/Scope: f5591f18-a6d0-4864-b08c-5a874023be..
			// order of the listed resources is not guaranteed so it must be checked one by one with substring
			Expect(cond.Message).To(ContainSubstring("Used by: "))
			Expect(cond.Message).To(ContainSubstring(fmt.Sprintf("%s/IpRange: %s", cloudcontrolv1beta1.GroupVersion.String(), subscriptionName)))
			Expect(cond.Message).To(ContainSubstring(fmt.Sprintf("%s/Scope: %s", cloudcontrolv1beta1.GroupVersion.String(), subscriptionName)))
		})

		By("When Subscription dependant labeled resources are deleted", func() {
			for _, x := range resources {
				Expect(Delete(infra.Ctx(), infra.KCP().Client(), x)).To(Succeed())
			}
			for _, x := range resources {
				Eventually(IsDeleted).
					WithArguments(infra.Ctx(), infra.KCP().Client(), x).
					Should(Succeed())
			}
		})

		By("Then KCP Subscription does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription).
				Should(Succeed())
		})

		By("// cleanup: Delete Secret and SecretBinding", func() {
			Expect(Delete(infra.Ctx(), infra.Garden().Client(), secret)).To(Succeed())
			Expect(Delete(infra.Ctx(), infra.Garden().Client(), secretBinding)).To(Succeed())
		})
	})

	It("Scenario: KCP Subscription GCP is created and deleted", func() {
		const (
			provider         = cloudcontrolv1beta1.ProviderGCP
			subscriptionName = "c05ccabd-39bb-4694-b36a-d2ae96463e02"
			gcpProjectId     = "gcp-project-id"
		)

		secret, secretBinding, subscription := commonInit(provider, subscriptionName, map[string][]byte{
			"project_id": []byte(gcpProjectId),
		})

		By("When KCP Subscription is created", func() {
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), subscription)).To(Succeed())
		})

		By("Then KCP Subscription is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		By("And Then KCP Subscription status.provider is set", func() {
			Expect(subscription.Status.Provider).To(Equal(provider))
		})

		By("And Then KCP Subscription status.subscriptionInfo.aws is set", func() {
			Expect(subscription.Status.SubscriptionInfo).NotTo(BeNil())
			Expect(subscription.Status.SubscriptionInfo.Gcp).NotTo(BeNil())
			Expect(subscription.Status.SubscriptionInfo.Gcp.Project).To(Equal(gcpProjectId))
		})

		By("And Then KCP Subscription has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(subscription, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		By("When KCP Subscription is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), subscription)).To(Succeed())
		})

		By("Then KCP Subscription does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription).
				Should(Succeed())
		})

		By("// cleanup: Delete Secret and SecretBinding", func() {
			Expect(Delete(infra.Ctx(), infra.Garden().Client(), secret)).To(Succeed())
			Expect(Delete(infra.Ctx(), infra.Garden().Client(), secretBinding)).To(Succeed())
		})
	})

	It("Scenario: KCP Subscription Azure is created and deleted", func() {
		const (
			provider            = cloudcontrolv1beta1.ProviderAzure
			subscriptionName    = "141a36e5-65ee-4896-8f61-e7bc783fbfb1"
			azureTenantId       = "d3337215-38ac-4360-ac3a-6bbc6b7cdb09"
			azureSubscriptionId = "c9fcebc6-2593-45d5-b269-f76e8271e880"
		)

		secret, secretBinding, subscription := commonInit(provider, subscriptionName, map[string][]byte{
			"subscriptionID": []byte(azureSubscriptionId),
			"tenantID":       []byte(azureTenantId),
		})

		By("When KCP Subscription is created", func() {
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), subscription)).To(Succeed())
		})

		By("Then KCP Subscription is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
		})

		By("And Then KCP Subscription status.provider is set", func() {
			Expect(subscription.Status.Provider).To(Equal(provider))
		})

		By("And Then KCP Subscription status.subscriptionInfo.aws is set", func() {
			Expect(subscription.Status.SubscriptionInfo).NotTo(BeNil())
			Expect(subscription.Status.SubscriptionInfo.Azure).NotTo(BeNil())
			Expect(subscription.Status.SubscriptionInfo.Azure.TenantId).To(Equal(azureTenantId))
			Expect(subscription.Status.SubscriptionInfo.Azure.SubscriptionId).To(Equal(azureSubscriptionId))
		})

		By("And Then KCP Subscription has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(subscription, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		By("When KCP Subscription is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), subscription)).To(Succeed())
		})

		By("Then KCP Subscription does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription).
				Should(Succeed())
		})

		By("// cleanup: Delete Secret and SecretBinding", func() {
			Expect(Delete(infra.Ctx(), infra.Garden().Client(), secret)).To(Succeed())
			Expect(Delete(infra.Ctx(), infra.Garden().Client(), secretBinding)).To(Succeed())
		})
	})
})
