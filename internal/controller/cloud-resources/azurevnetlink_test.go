package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR AzureVNetLink", func() {

	It("Scenario: SKR AzureVNetLink is created then deleted", func() {
		const (
			remoteSubscription  = "cb9adcc4-97f4-4216-a16f-d2afe35d5077"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
		)
		azureVNetLink := &cloudresourcesv1beta1.AzureVNetLink{}

		remoteVnetId := util.NewVirtualNetworkResourceId(remoteSubscription, remoteResourceGroup, remoteVnetName).String()

		By("When AzureVNetLink is created", func() {
			Eventually(CreateAzureVNetLink).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureVNetLink,
					WithName("dea5b922-f7be-404c-9f69-72dfce914bd2"),
					WithAzureRemoteVNetLinkName("91953457-3728-4e40-b874-ac4717f9d43e"),
					WithAzureRemotePrivateDnsZone(remoteVnetId),
				).Should(Succeed())
		})

		By("Then AzureVNetLink has status.ID", func() {
			// load SKR AzureVNetLink to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureVNetLink,
					NewObjActions(),
					AssertAzureVNetPeeringHasId(),
				).
				Should(Succeed(), "expected AzureVNetLink to get status.Id, but it didn't")
		})

		kcpAzureVNetLink := &cloudcontrolv1beta1.AzureVNetLink{}

		By("Then KCP AzureVNetLink is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpAzureVNetLink,
					NewObjActions(WithName(azureVNetLink.Status.Id)),
				).
				Should(Succeed(), "failed to load KCP AzureVNetLink")
		})

		By("And Then KCP AzureVNetLink has annotations", func() {
			Expect(kcpAzureVNetLink.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))
			Expect(kcpAzureVNetLink.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(azureVNetLink.Name))
			Expect(kcpAzureVNetLink.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(azureVNetLink.Namespace))
		})

		By("When KCP AzureVNetLink is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					kcpAzureVNetLink,
					WithState(cloudcontrolv1beta1.VirtualNetworkLinkStateCompleted),
					WithConditions(KcpReadyCondition())).
				Should(Succeed(), "failed to update status on KCP AzureVNetLink")
		})

		By("Then SKR AzureVNetLink is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureVNetLink,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState(cloudcontrolv1beta1.VirtualNetworkLinkStateCompleted)).
				Should(Succeed(), "expect SKR AzureVNetLink to be Ready, but it didn't")
		})

		By("When SKR AzureVNetLink is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), azureVNetLink).
				Should(Succeed(), "failed to delete SKR AzureVNetLink")
		})

		By("Then KCP AzureVNetLink does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpAzureVNetLink, WithName(azureVNetLink.Status.Id)).
				Should(Succeed(), "failed to delete KCP AzureVNetLink")
		})
	})

})
