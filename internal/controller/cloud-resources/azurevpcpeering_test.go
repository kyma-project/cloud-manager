package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/kcp/provider/azure/util"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR AzureVpcPeering", func() {

	It("Scenario: SKR AzureVpcPeering is created then deleted", func() {
		const (
			remoteSubscription  = "2d11eacb-1667-48c6-a455-12de3757db03"
			remoteResourceGroup = "MyResourceGroup"
			remoteVnetName      = "MyVnet"
		)
		azureVpcPeering := &cloudresourcesv1beta1.AzureVpcPeering{}

		remoteVnetId := util.NewVirtualNetworkResourceId(remoteSubscription, remoteResourceGroup, remoteVnetName).String()

		By("When AzureVpcPeering is created", func() {
			Eventually(CreateAzureVpcPeering).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), azureVpcPeering,
					WithName("0d247cc0-dffb-40c1-a9f3-fbd3b4591f9f"),
					WithAzureRemotePeeringName("b5d42fd8-2623-49d6-8792-1fb50f7e1fb6"),
					WithAzureRemoteVnet(remoteVnetId),
				).Should(Succeed())
		})

		By("Then AzureVpcPeering has status.ID", func() {
			// load SKR AzureVpcPeering to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureVpcPeering,
					NewObjActions(),
					AssertAzureVpcPeeringHasId(),
				).
				Should(Succeed(), "expected AzureVpcPeering to get status.Id, but it didn't")
		})

		remoteNetwork := &cloudcontrolv1beta1.Network{}

		By("Then KCP remote Network is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					remoteNetwork,
					NewObjActions(WithName(azureVpcPeering.Status.Id)),
				).
				Should(Succeed(), "failed to load remote Network")
		})

		By("And Than KCP remote Network has AzureNetworkReference", func() {
			Expect(remoteNetwork.Spec.Network.Reference.Azure.SubscriptionId).To(Equal(remoteSubscription))
			Expect(remoteNetwork.Spec.Network.Reference.Azure.NetworkName).To(Equal(remoteVnetName))
			Expect(remoteNetwork.Spec.Network.Reference.Azure.ResourceGroup).To(Equal(remoteResourceGroup))
		})

		vpcPeering := &cloudcontrolv1beta1.VpcPeering{}

		By("Then KCP VpcPeering is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					vpcPeering,
					NewObjActions(WithName(azureVpcPeering.Status.Id)),
				).
				Should(Succeed(), "failed to load KCP VpcPeering")
		})

		By("And Then KCP VpcPeering has RemoteNetwork object reference", func() {
			Expect(vpcPeering.Spec.Details.RemoteNetwork.Name).To(Equal(azureVpcPeering.Status.Id))
			Expect(vpcPeering.Spec.Details.RemoteNetwork.Namespace).To(Equal(DefaultKcpNamespace))
		})

		By("And Then KCP VpcPeering has LocalNetwork object reference", func() {
			Expect(vpcPeering.Spec.Details.LocalNetwork.Name).To(Equal(common.KymaNetworkCommonName(vpcPeering.Spec.Scope.Name)))
			Expect(vpcPeering.Spec.Details.LocalNetwork.Namespace).To(Equal(DefaultKcpNamespace))
		})

		By("And Then KCP VpcPeering has annotations", func() {
			Expect(vpcPeering.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))
			Expect(vpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(azureVpcPeering.Name))
			Expect(vpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(azureVpcPeering.Namespace))
		})

		By("When KCP VpcPeering is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					vpcPeering,
					WithState(cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected),
					WithConditions(KcpReadyCondition()))
		})

		By("Then SKR AzureVpcPeering is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					azureVpcPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState(cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected),
				)
		})

		By("When SKR AzureVpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), azureVpcPeering).
				Should(Succeed(), "failed to delete SKR AzureVpcPeering")
		})

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcPeering, WithName(azureVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP VpcPeering")
		})

		By("Then KCP remote Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNetwork, WithName(azureVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP remote Network")
		})

	})

})
