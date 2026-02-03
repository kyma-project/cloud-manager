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
	"github.com/gophercloud/gophercloud/v2/openstack/networking/v2/networks"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
)

var _ = Describe("Feature: VpcNetwork", func() {

	It("Scenario: VpcNetwork SAP is created", func() {
		const subscriptionName = "7227bef8-3edf-4d9a-8bc5-e56f7a068f2d"
		const vpcNetworkName = "08a9f82b-b0ab-4036-9b07-6a588b2ee711"

		subscription := &cloudcontrolv1beta1.Subscription{}
		var vpcNetwork *cloudcontrolv1beta1.VpcNetwork

		sapMock := infra.SapMock().NewProject()

		By("Given OpenStack external network exists", func() {
			Expect(SapInfraInitializeExternalNetworks(infra.Ctx(), sapMock)).
				To(Succeed())
		})

		By("Given SAP Subscription exists", func() {
			kcpsubscription.Ignore.AddName(subscriptionName)
			Expect(
				CreateSubscription(infra.Ctx(), infra, subscription,
					WithName(subscriptionName),
					WithSubscriptionSpecGarden("binding-name")),
			).To(Succeed())

			Expect(
				SubscriptionPatchStatusReadyOpenStack(infra.Ctx(), infra, subscription, sapMock.DomainName(), sapMock.ProjectName()),
			).To(Succeed())
		})

		By("When VpcNetwork is created", func() {
			vpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithRegion(sapMock.RegionName()).
				WithSubscription(subscriptionName).
				WithCidrBlocks("10.250.0.0/16").
				Build()

			Expect(
				CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork, WithName(vpcNetworkName)),
			).To(Succeed())
		})

		By("Then VpcNetwork is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork, NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed())
			cond := meta.FindStatusCondition(*vpcNetwork.Conditions(), cloudcontrolv1beta1.ReasonReady)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Reason).To(Equal(cloudcontrolv1beta1.ReasonProvisioned))
		})

		By("Then VpcNetwork has subscription label", func() {
			Expect(vpcNetwork.Labels[cloudcontrolv1beta1.SubscriptionLabel]).To(Equal(subscriptionName))
		})

		By("Then VpcNetwork has provider label", func() {
			Expect(vpcNetwork.Labels[cloudcontrolv1beta1.SubscriptionLabelProvider]).To(Equal(string(cloudcontrolv1beta1.ProviderOpenStack)))
		})

		By("Then VpcNetwork status has vpcID", func() {
			Expect(vpcNetwork.Status.Identifiers.Vpc).NotTo(BeEmpty())
		})

		var vpc *networks.Network

		By("Then SAP Internal Network exists", func() {
			aVpc, err := sapMock.GetNetwork(infra.Ctx(), vpcNetwork.Status.Identifiers.Vpc)
			Expect(err).ToNot(HaveOccurred())
			Expect(aVpc).ToNot(BeNil())
			vpc = aVpc
		})

		// DELETE ===============================================

		By("When KCP VpcNetwork is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())
		})

		By("Then KCP VpcNetwork does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork).
				Should(Succeed())
		})

		By("Then SAP Internal Network does not exist", func() {
			aVpc, err := sapMock.GetNetwork(infra.Ctx(), vpc.ID)
			Expect(err).ToNot(HaveOccurred())
			Expect(aVpc).To(BeNil())
		})

		By("// cleanup: delete Subscription", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), subscription)).To(Succeed())
		})
	})

})
