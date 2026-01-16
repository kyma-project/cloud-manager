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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: VpcNetwork", func() {

	It("Scenario: VpcNetwork Azure is created", func() {
		const tenantId = "55504913-70e5-4cf1-9312-af1b16146f2d"
		const subscriptionId = "894d9fab-9a89-4c01-b535-fd222ceee970"
		const name = "4903778f-9e18-474a-a3d7-b4fadbae03bf"
		const region = "eastus2"

		subscription := &cloudcontrolv1beta1.Subscription{}
		var vpcNetwork *cloudcontrolv1beta1.VpcNetwork

		azureMock := infra.AzureMock().MockConfigs(subscriptionId, tenantId)

		By("Given Azure Subscription exists", func() {
			kcpsubscription.Ignore.AddName(name)
			Expect(
				CreateSubscription(infra.Ctx(), infra, subscription,
					WithName(name),
					WithSubscriptionSpecGarden("binding-name")),
			).To(Succeed())

			Expect(
				SubscriptionPatchStatusReadyAzure(infra.Ctx(), infra, subscription, tenantId, subscriptionId),
			).To(Succeed())
		})

		By("When VpcNetwork is created", func() {
			vpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithName(name).
				WithRegion(region).
				WithSubscription(name).
				WithCidrBlocks("10.250.0.0/16").
				Build()

			Expect(
				CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork),
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
			Expect(vpcNetwork.Labels[cloudcontrolv1beta1.SubscriptionLabel]).To(Equal(name))
		})

		By("Then VpcNetwork has provider label", func() {
			Expect(vpcNetwork.Labels[cloudcontrolv1beta1.SubscriptionLabelProvider]).To(Equal(string(cloudcontrolv1beta1.ProviderAzure)))
		})

		var resourceGroupID azureutil.ResourceDetails

		By("Then VpcNetwork status has resource group ID", func() {
			Expect(vpcNetwork.Status.Identifiers.Vpc).NotTo(BeEmpty())
			rd, err := azureutil.ParseResourceID(vpcNetwork.Status.Identifiers.ResourceGroup)
			Expect(err).NotTo(HaveOccurred())
			resourceGroupID = rd
		})

		var vpcID azureutil.ResourceDetails

		By("Then VpcNetwork status has vpc network ID", func() {
			Expect(vpcNetwork.Status.Identifiers.Vpc).NotTo(BeEmpty())
			rd, err := azureutil.ParseResourceID(vpcNetwork.Status.Identifiers.Vpc)
			Expect(err).NotTo(HaveOccurred())
			vpcID = rd
			Expect(resourceGroupID.ResourceGroup).To(Equal(vpcID.ResourceGroup))
		})

		var azureVirtualNetwork *armnetwork.VirtualNetwork

		By("Then Azure VPC Network exists", func() {
			vpc, err := azureMock.GetNetwork(infra.Ctx(), vpcID.ResourceGroup, vpcID.ResourceName)
			Expect(err).ToNot(HaveOccurred())
			Expect(vpc).ToNot(BeNil())
			azureVirtualNetwork = vpc
		})

		By("Then Azure VPC Network has correct CIDR block", func() {
			Expect(pie.Map(azureVirtualNetwork.Properties.AddressSpace.AddressPrefixes, func(x *string) string {
				return ptr.Deref(x, "")
			})).To(Equal(vpcNetwork.Status.CidrBlocks))
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

		By("Then Azure VPC Network does not exist", func() {
			vpc, err := azureMock.GetNetwork(infra.Ctx(), vpcID.ResourceGroup, vpcID.ResourceName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(vpc).To(BeNil())
		})

		By("Then Azure VPC Network does not exist", func() {
			rg, err := azureMock.GetResourceGroup(infra.Ctx(), resourceGroupID.ResourceGroup)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
			Expect(rg).To(BeNil())
		})

		By("// cleanup: delete Subscription", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), subscription)).To(Succeed())
		})
	})

})
