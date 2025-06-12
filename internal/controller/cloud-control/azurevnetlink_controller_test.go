package cloudcontrol

import (
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	azureutil "github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP AzureVNetLink", func() {

	It("Scenario: KCP AzureVNetLink is created and deleted", func() {
		const (
			kymaName                     = "56d6f6ba-083e-4b1a-aeff-813c517722fb"
			kcpAzureVNetLinkName         = "58935fd4-9f0c-4203-911f-dba920d822ea"
			remoteSubscription           = "1c65f69a-dba5-4979-bd16-a2a7e22a9e3d"
			remoteResourceGroup          = "MyResourceGroup"
			remotePrivateDnsZoneName     = "MyPrivateDnsZone"
			remoteVirtualPrivateLinkName = "example-com"
		)

		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Scope exists", func() {
			// Tell Scope reconciler to ignore this kymaName
			kcpscope.Ignore.AddName(kymaName)

			Eventually(CreateScopeAzure).
				WithArguments(infra.Ctx(), infra, scope, WithName(kymaName)).
				Should(Succeed())
		})

		//localResourceGroupName := scope.Spec.Scope.Azure.VpcNetwork
		//localVirtualNetworkName := scope.Spec.Scope.Azure.VpcNetwork

		// azureMockLocal := infra.AzureMock().MockConfigs(scope.Spec.Scope.Azure.SubscriptionId, scope.Spec.Scope.Azure.TenantId)

		azureMockRemote := infra.AzureMock().MockConfigs(remoteSubscription, scope.Spec.Scope.Azure.TenantId)

		By("And Given that remote private DNS zone exists", func() {
			err := azureMockRemote.CreatePrivateDnsZone(infra.Ctx(), remoteResourceGroup, remotePrivateDnsZoneName, map[string]string{kymaName: kymaName})
			Expect(err).NotTo(HaveOccurred())
		})

		var azureVNetLink *cloudcontrolv1beta1.AzureVNetLink

		By("When KCP AzureVnetLink is created", func() {
			azureVNetLink = (&cloudcontrolv1beta1.AzureVNetLinkBuilder{}).
				WithScope(kymaName).
				WithRemotePrivateDnsZone(azureutil.NewPrivateDnsZoneResourceId(remoteSubscription, remoteResourceGroup, remotePrivateDnsZoneName).String()).
				WithRemoteVirtualPrivateLinkName(remoteVirtualPrivateLinkName).
				Build()

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.KCP().Client(), azureVNetLink,
					WithName(kcpAzureVNetLinkName),
				).
				Should(Succeed())
		})

		By("Then KCP AzureVnetLink has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), azureVNetLink,
					NewObjActions(),
					HaveFinalizer(api.CommonFinalizerDeletionHook),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
				).
				Should(Succeed())
		})

		By("Then VirtualNetworkLink exist", func() {
			link, err := azureMockRemote.GetVirtualNetworkLink(infra.Ctx(), remoteResourceGroup, remotePrivateDnsZoneName, remoteVirtualPrivateLinkName)
			Expect(err).NotTo(HaveOccurred())
			Expect(link).NotTo(BeNil())
		})

		// DELETE

		By("When KCP AzureVnetLink is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), azureVNetLink).
				Should(Succeed(), "failed deleting AzureVnetLink")
		})

		By("Then KCP AzureVnetLink does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), azureVNetLink).
				Should(Succeed(), "expected AzureVnetLink not to exist (be deleted), but it still exists")
		})

		By("And Then VirtualNetworkLink does not exist", func() {
			link, err := azureMockRemote.GetVirtualNetworkLink(infra.Ctx(), remoteResourceGroup, remotePrivateDnsZoneName, remoteVirtualPrivateLinkName)
			Expect(err).To(HaveOccurred())
			Expect(link).To(BeNil())
		})

		By("// cleanup: Scope", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed())
		})

	})

})
