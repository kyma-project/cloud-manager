package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR AwsVpcPeering", func() {

	It("Scenario: SKR AwsVpcPeering is created then deleted", func() {
		awsVpcPeering := &cloudresourcesv1beta1.AwsVpcPeering{}

		const (
			remoteAccountId = "444455556666"
			remoteRegion    = "eu-west-1"
			remoteVpcId     = "vpc-2c41e43fcd5340f8f"
		)

		By("When AwsVpcPeering is created", func() {
			Eventually(CreateAwsVpcPeering).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsVpcPeering,
					WithName("2c371635-5256-46c6-aaf1-e49c439d985c"),
					WithAwsRemoteAccountId(remoteAccountId),
					WithAwsRemoteRegion(remoteRegion),
					WithAwsRemoteVpcId(remoteVpcId),
				).Should(Succeed())
		})

		By("Then KCP VpcPeering is created", func() {
			// load SKR AwsVpcPeering to get ID
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsVpcPeering,
					NewObjActions(),
					AssertAwsVpcPeeringHasId(),
				).
				Should(Succeed(), "expected AwsVpcPeering to get status.Id, but it didn't")
		})

		remoteNetwork := &cloudcontrolv1beta1.Network{}

		By("Then KCP remote Network is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					remoteNetwork,
					NewObjActions(WithName(awsVpcPeering.Status.Id)),
				).
				Should(Succeed(), "failed to load remote Network")
		})

		By("And Than KCP remote Network has AwsNetworkReference", func() {
			Expect(remoteNetwork.Spec.Network.Reference.Aws.AwsAccountId).To(Equal(remoteAccountId))
			Expect(remoteNetwork.Spec.Network.Reference.Aws.Region).To(Equal(remoteRegion))
			Expect(remoteNetwork.Spec.Network.Reference.Aws.VpcId).To(Equal(remoteVpcId))
		})

		kcpVpcPeering := &cloudcontrolv1beta1.VpcPeering{}

		By("Then AwsVpcPeering is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpVpcPeering,
					NewObjActions(WithName(awsVpcPeering.Status.Id)),
				).
				Should(Succeed(), "failed to load KCP VpcPeering")
		})

		By("And Then KCP VpcPeering has RemoteNetwork object reference", func() {
			Expect(kcpVpcPeering.Spec.Details.RemoteNetwork.Name).To(Equal(awsVpcPeering.Status.Id))
			Expect(kcpVpcPeering.Spec.Details.RemoteNetwork.Namespace).To(Equal(DefaultKcpNamespace))
		})

		By("And Then KCP VpcPeering has LocalNetwork object reference", func() {
			Expect(kcpVpcPeering.Spec.Details.LocalNetwork.Name).To(Equal(common.KcpNetworkKymaCommonName(kcpVpcPeering.Spec.Scope.Name)))
			Expect(kcpVpcPeering.Spec.Details.LocalNetwork.Namespace).To(Equal(DefaultKcpNamespace))
		})

		By("And Then KCP VpcPeering has annotations", func() {
			Expect(kcpVpcPeering.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))
			Expect(kcpVpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(awsVpcPeering.Name))
			Expect(kcpVpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(awsVpcPeering.Namespace))
		})

		By("When KCP VpcPeering is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					kcpVpcPeering,
					WithState(cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected),
					WithConditions(KcpReadyCondition()))
		})

		By("Then SKR AwsVpcPeering is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					awsVpcPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady),
					HavingState(cloudcontrolv1beta1.VirtualNetworkPeeringStateConnected),
				)
		})

		By("When SKR AwsVpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), awsVpcPeering).
				Should(Succeed(), "failed to delete SKR AwsVpcPeering")
		})

		By("Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcPeering, WithName(awsVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP VpcPeering")
		})

		By("And Then KCP remote Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNetwork, WithName(awsVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP remote Network")
		})

	})

})
