package cloudcontrol

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/elliotchance/pie/v2"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	commongardener "github.com/kyma-project/cloud-manager/pkg/common/gardener"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: VpcNetwork", func() {

	It("Scenario: Azure VpcNetwork type Kyma is created and deleted", func() {
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
			Expect(vpcNetwork.Status.Identifiers.ResourceGroup).NotTo(BeEmpty())
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

		By("Then VpcNetwork status has name", func() {
			Expect(vpcNetwork.Status.Identifiers.Name).NotTo(BeEmpty())
			Expect(vpcNetwork.Status.Identifiers.Name).To(Equal("kyma-default-" + name))
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

		By("Then Azure Resource Group does not exist", func() {
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

	It("Scenario: Azure VpcNetwork type Gardener is created and deleted", func() {
		const tenantId = "19a0c7b5-5071-4349-88a2-c8db6b1558e2"
		const subscriptionId = "af762091-bff6-468a-bbed-8b7dec401f8e"
		const name = "f4347d91-6a2b-42c0-8ed1-3adfdf450d4f"
		const shootName = "x-06adb5e"
		const region = "eastus2"
		const bindingName = "binding-06adb5e"

		var vpcNetworkName string

		azureMock := infra.AzureMock().MockConfigs(subscriptionId, tenantId)

		subscription := &cloudcontrolv1beta1.Subscription{}
		vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

		var azureResourceGroup *armresources.ResourceGroup
		var azureVirtualNetwork *armnetwork.VirtualNetwork

		_ = azureResourceGroup
		_ = azureVirtualNetwork

		By("Given shoot name and namespace", func() {
			shootNamespace, err := commongardener.DefaultGardenerNamespaceProvider().GetGardenerNamespace(infra.Ctx(), infra.KCP().Client())
			Expect(err).ToNot(HaveOccurred())
			vpcNetworkName = common.GardenerVpcName(shootNamespace, shootName)
		})

		By("Given Azure Subscription exists", func() {
			kcpsubscription.Ignore.AddName(name)
			Expect(
				CreateSubscription(infra.Ctx(), infra, subscription,
					WithName(name),
					WithLabels(map[string]string{
						// IMPORTANT! The Subscription reconciler would put this label, but since added to ignore, we must set it
						cloudcontrolv1beta1.SubscriptionLabelBindingName: bindingName,
					}),
					WithSubscriptionSpecGarden(bindingName)),
			).To(Succeed())

			Expect(
				SubscriptionPatchStatusReadyAzure(infra.Ctx(), infra, subscription, tenantId, subscriptionId),
			).To(Succeed())
		})

		By("When VpcNetwork is created", func() {
			vpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithType(cloudcontrolv1beta1.VpcNetworkTypeGardener).
				WithName(name).
				WithVpcNetworkName(new(vpcNetworkName)).
				WithRegion(region).
				WithSubscription(name).
				WithCidrBlocks("10.250.0.0/16").
				Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())
		})

		By("Then VpcNetwork has provider error observing resource group", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), vpcNetwork, NewObjActions(),
					HavingCondition(
						cloudcontrolv1beta1.ConditionTypeReady,
						metav1.ConditionFalse,
						cloudcontrolv1beta1.ReasonProviderError,
						"Error observing resource group",
					),
				).
				Should(Succeed())
		})

		By("When Azure ResourceGroup is created", func() {
			rg, err := azureMock.CreateResourceGroup(infra.Ctx(), vpcNetworkName, region, nil)
			Expect(err).ToNot(HaveOccurred())
			azureResourceGroup = rg
		})

		By("Then VpcNetwork has provider error observing virtual network", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), vpcNetwork, NewObjActions(),
					HavingCondition(
						cloudcontrolv1beta1.ConditionTypeReady,
						metav1.ConditionFalse,
						cloudcontrolv1beta1.ReasonProviderError,
						"Error observing virtual network",
					),
				).
				Should(Succeed())
		})

		By("When Azure Virtual Network is created", func() {
			poller, err := azureMock.CreateOrUpdateNetwork(infra.Ctx(), vpcNetworkName, vpcNetworkName, armnetwork.VirtualNetwork{
				Location: new(region),
				Properties: &armnetwork.VirtualNetworkPropertiesFormat{
					AddressSpace: &armnetwork.AddressSpace{
						AddressPrefixes: []*string{new("10.250.0.0/16")},
					},
				},
			}, nil)
			Expect(err).ToNot(HaveOccurred())
			resp, err := poller.PollUntilDone(infra.Ctx(), nil)
			Expect(err).ToNot(HaveOccurred())
			azureVirtualNetwork = &resp.VirtualNetwork
		})

		By("Then VpcNetwork is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), vpcNetwork, NewObjActions(),
					HavingCondition(
						cloudcontrolv1beta1.ConditionTypeReady,
						metav1.ConditionTrue,
						cloudcontrolv1beta1.ReasonProvisioned,
						"",
					),
				).
				Should(Succeed())
		})

		By("And Then VpcNetwork has status.identifiers.name", func() {
			Expect(vpcNetwork.Status.Identifiers.Name).ToNot(BeEmpty())
			Expect(vpcNetwork.Status.Identifiers.Name).To(Equal(vpcNetworkName))
		})

		By("And Then VpcNetwork has status.identifiers.vpc", func() {
			Expect(vpcNetwork.Status.Identifiers.Vpc).ToNot(BeEmpty())
			Expect(vpcNetwork.Status.Identifiers.Vpc).To(Equal(ptr.Deref(azureVirtualNetwork.ID, "")))
		})

		By("And Then VpcNetwork has status.identifiers.resourceGroup", func() {
			Expect(vpcNetwork.Status.Identifiers.Vpc).ToNot(BeEmpty())
			Expect(vpcNetwork.Status.Identifiers.Vpc).To(Equal(ptr.Deref(azureVirtualNetwork.ID, "")))
		})

		// Delete

		By("When KCP VpcNetwork is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())
		})

		By("Then KCP VpcNetwork does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork).
				Should(Succeed())
		})
	})
})
