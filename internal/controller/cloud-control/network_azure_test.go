package cloudcontrol

import (
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v5"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	azurecommon "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/common"
	azuremeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/meta"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: KCP Azure managed Network", func() {

	It("Scenario: Managed Azure network is created and deleted", func() {

		kymaName := "8515c338-70ec-41f6-8ac8-639d636daf1b"
		scope := &cloudcontrolv1beta1.Scope{}
		netObjName := kymaName + "--cm" // !important to be CM network

		var net *cloudcontrolv1beta1.Network

		By("Given Scope exists", func() {
			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		azureMock := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)
		expectedResourceGroupName := azurecommon.AzureCloudManagerResourceGroupName(scope.Spec.Scope.Azure.VpcNetwork)

		By("When managed CM KCP Network is created", func() {
			net = cloudcontrolv1beta1.NewNetworkBuilder().WithManagedNetwork().Build()
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net, WithName(netObjName), WithScope(kymaName)).
				Should(Succeed())
		})

		By("Then KCP Network state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net, NewObjActions(), HavingState(string(cloudcontrolv1beta1.StateReady))).
				Should(Succeed())
		})

		By("And Then KCP Network has Ready condition", func() {
			cond := meta.FindStatusCondition(*net.Conditions(), cloudcontrolv1beta1.ConditionTypeReady)
			Expect(cond).ToNot(BeNil())
			Expect(cond.Status).To(Equal(metav1.ConditionTrue))
		})

		By("And Then KCP Network has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(net, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		By("And Then Network status reference is set", func() {
			Expect(net.Status.Network).NotTo(BeNil())
			Expect(net.Status.Network.Azure).NotTo(BeNil())
		})

		By("And Then Network status reference tenant equals to Scope tenantId", func() {
			Expect(net.Status.Network.Azure.TenantId).To(Equal(scope.Spec.Scope.Azure.TenantId))
		})

		By("And Then Network status reference subscription equals to Scope subscriptionId", func() {
			Expect(net.Status.Network.Azure.SubscriptionId).To(Equal(scope.Spec.Scope.Azure.SubscriptionId))
		})

		By("And Then Network status reference resourceGroup equals to predefined CloudManager resource group name", func() {
			Expect(net.Status.Network.Azure.ResourceGroup).To(Equal(expectedResourceGroupName))
		})

		By("And Then Network status reference network name equals to CM Network name", func() {
			Expect(net.Status.Network.Azure.NetworkName).To(Equal(expectedResourceGroupName))
		})

		var azureResourceGroup *armresources.ResourceGroup

		By("And Then Azure ResourceGroup is created", func() {
			rd, err := azureMock.GetResourceGroup(infra.Ctx(), expectedResourceGroupName)
			Expect(err).NotTo(HaveOccurred())
			Expect(rd).NotTo(BeNil())
			Expect(ptr.Deref(rd.Name, "")).To(Equal(expectedResourceGroupName))
			azureResourceGroup = rd
		})

		By("And Then Azure ResourceGroup location equals to Scope region", func() {
			Expect(ptr.Deref(azureResourceGroup.Location, "")).To(Equal(scope.Spec.Region))
		})

		var azureVNet *armnetwork.VirtualNetwork

		By("And Then Azure CM VNet is created", func() {
			vnet, err := azureMock.GetNetwork(infra.Ctx(), expectedResourceGroupName, expectedResourceGroupName)
			Expect(err).NotTo(HaveOccurred())
			Expect(vnet).NotTo(BeNil())
			azureVNet = vnet
		})

		By("And Then Azure VNet location equals to Scope region", func() {
			Expect(ptr.Deref(azureVNet.Location, "")).To(Equal(scope.Spec.Region))
		})

		By("And Then Azure VNet address space equals to default CloudManager CIDR", func() {
			Expect(azureVNet.Properties).NotTo(BeNil())
			Expect(azureVNet.Properties.AddressSpace).NotTo(BeNil())
			Expect(azureVNet.Properties.AddressSpace.AddressPrefixes).NotTo(BeNil())
			Expect(azureVNet.Properties.AddressSpace.AddressPrefixes).To(HaveLen(1))
			Expect(ptr.Deref(azureVNet.Properties.AddressSpace.AddressPrefixes[0], "")).To(Equal(common.DefaultCloudManagerCidr))
		})

		// Delete ======================================================================================================

		By("When KCP Network is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net).
				Should(Succeed())
		})

		By("Then KCP Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), net).
				Should(Succeed())
		})

		By("And Then Azure VNet does not exist", func() {
			_, err := azureMock.GetNetwork(infra.Ctx(), expectedResourceGroupName, expectedResourceGroupName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
		})

		By("And Then Azure ResourceGroup does not exist", func() {
			_, err := azureMock.GetResourceGroup(infra.Ctx(), expectedResourceGroupName)
			Expect(err).To(HaveOccurred())
			Expect(azuremeta.IsNotFound(err)).To(BeTrue())
		})
	})

})
