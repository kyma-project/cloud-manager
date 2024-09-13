package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
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
					WithName("skr-azure-vpcpeering"),
					WithAzureRemotePeeringName("peering-to-my-vnet"),
					WithAzureRemoteVnet(remoteVnetId),
				).Should(Succeed())
		})

		vpcPeering := &cloudcontrolv1beta1.VpcPeering{}

		By("Then KCP VpcPeering is created", func() {
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

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					vpcPeering,
					NewObjActions(WithName(azureVpcPeering.Status.Id)),
				).
				Should(Succeed(), "failed to load KCP VpcPeering")

			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcPeering, AddFinalizer(cloudcontrolv1beta1.FinalizerName)).
				Should(Succeed(), "failed adding finalizer on KCP VpcPeering")
		})

		By("And Then KCP VpcPeering has annotation cloud-manager.kyma-project.io/kymaName", func() {
			Expect(vpcPeering.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))
		})

		By("And Then KCP VpcPeering has annotation cloud-manager.kyma-project.io/remoteName", func() {
			Expect(vpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(azureVpcPeering.Name))
		})

		By("And Then KCP VpcPeering has annotation cloud-manager.kyma-project.io/remoteNamespace", func() {
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

		By("Then KCP VpcPeering is marked for deletion", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcPeering, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed(), "expected KCP VpcPeering to be marked for deletion")
		})

		By("When KCP VpcPeering finalizer is removed and it is deleted", func() {
			Eventually(Update).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcPeering, RemoveFinalizer(cloudcontrolv1beta1.FinalizerName)).
				Should(Succeed(), "failed removing finalizer on KCP VpcPeering")
		})

		By("Then KCP VpcPeering is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcPeering, WithName(azureVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP VpcPeering")
		})

		remoteNetwork := &cloudcontrolv1beta1.Network{}

		By("Then KCP remote Network is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNetwork, WithName(azureVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP remote Network")
		})

	})

})
