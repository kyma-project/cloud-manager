package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR AwsVpcPeering", func() {

	It("Scenario: SKR AwsVpcPeering is created then deleted", func() {
		awsVpcPeering := &cloudresourcesv1beta1.AwsVpcPeering{}

		By("When AwsVpcPeering is created", func() {
			Eventually(CreateAwsVpcPeering).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), awsVpcPeering,
					WithName("vpc-2c41e43fcd5340f8f"),
					WithAwsRemoteAccountId("444455556666"),
					WithAwsRemoteRegion("eu-west-1"),
					WithAwsRemoteVpcId("vpc-2c41e43fcd5340f8f"),
				).Should(Succeed())
		})

		vpcPeering := &cloudcontrolv1beta1.VpcPeering{}

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

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					vpcPeering,
					NewObjActions(WithName(awsVpcPeering.Status.Id)),
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
			Expect(vpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(awsVpcPeering.Name))
		})

		By("And Then KCP VpcPeering has annotation cloud-manager.kyma-project.io/remoteNamespace", func() {
			Expect(vpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(awsVpcPeering.Namespace))
		})

		By("When KCP VpcPeering is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					vpcPeering,
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
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcPeering, WithName(awsVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP VpcPeering")
		})

		remoteNetwork := &cloudcontrolv1beta1.Network{}

		By("And Then KCP remote Network is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNetwork, WithName(awsVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP remote Network")
		})

	})

})
