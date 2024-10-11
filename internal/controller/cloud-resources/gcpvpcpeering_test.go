package cloudresources

import (
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: SKR GcpVpcPeering", func() {

	It("Scenario: Creating SKR GcpVpcPeering and then deleting it", func() {

		const (
			remoteProject     = "my-gcp-project"
			remoteVpc         = "my-gcp-vpc"
			remotePeeringName = "peering-kyma-dev-to-my-gcp-project"
			importCustomRoute = false
		)

		gcpVpcPeering := &cloudresourcesv1beta1.GcpVpcPeering{}
		By("When SKR GcpVpcPeering is created", func() {
			Eventually(CreateObj).
				WithArguments(
					infra.Ctx(), infra.SKR().Client(), gcpVpcPeering,
					WithName("peering-kyma-dev-to-my-gcp-project"),
					WithGcpPeeringName(remotePeeringName),
					WithGcpRemoteProject(remoteProject),
					WithGcpRemoteVpc(remoteVpc),
					WithImportCustomRoute(importCustomRoute),
				).Should(Succeed())
		})

		By("Then SKR GcpVpcPeering will create a unique status.id", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpVpcPeering,
					NewObjActions(),
					HavingGcpVpcPeeringStatusId(),
				).
				Should(Succeed(), "failed to load SKR GcpVpcPeering with unique id")
		})

		remoteNetwork := &cloudcontrolv1beta1.Network{}

		By("And Then KCP remote Network is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					remoteNetwork,
					NewObjActions(WithName(gcpVpcPeering.Status.Id)),
				).
				Should(Succeed(), "failed to load remote Network")
		})

		By("When KCP remote Network is ready", func() {
			Eventually(UpdateStatus).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					remoteNetwork,
					WithConditions(KcpReadyCondition()),
				).
				Should(Succeed(), "failed to load remote Network")
		})

		By("Then KCP remote Network matches the details on the Gcp Vpc Peering", func() {
			Expect(remoteNetwork.Spec.Network.Reference.Gcp.GcpProject).To(Equal(remoteProject))
			Expect(remoteNetwork.Spec.Network.Reference.Gcp.NetworkName).To(Equal(remoteVpc))
		})

		kcpVpcPeering := &cloudcontrolv1beta1.VpcPeering{}
		By("And Then KCP VpcPeering is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.KCP().Client(),
					kcpVpcPeering,
					NewObjActions(WithName(gcpVpcPeering.Status.Id)),
				).
				Should(Succeed(), "KCP GcpVpcPeering does not exist")
		})

		By("And Then KCP VpcPeering has RemoteNetwork object reference", func() {
			Expect(kcpVpcPeering.Spec.Details.RemoteNetwork.Name).To(Equal(gcpVpcPeering.Status.Id))
			Expect(kcpVpcPeering.Spec.Details.RemoteNetwork.Namespace).To(Equal(DefaultKcpNamespace))
		})

		By("And Then KCP VpcPeering has LocalNetwork object reference", func() {
			Expect(kcpVpcPeering.Spec.Details.LocalNetwork.Name).To(Equal(common.KcpNetworkKymaCommonName(kcpVpcPeering.Spec.Scope.Name)))
			Expect(kcpVpcPeering.Spec.Details.LocalNetwork.Namespace).To(Equal(DefaultKcpNamespace))
		})

		By("And Then KCP VpcPeering has annotations", func() {
			Expect(kcpVpcPeering.Annotations[cloudcontrolv1beta1.LabelKymaName]).To(Equal(infra.SkrKymaRef().Name))
			Expect(kcpVpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteName]).To(Equal(gcpVpcPeering.Name))
			Expect(kcpVpcPeering.Annotations[cloudcontrolv1beta1.LabelRemoteNamespace]).To(Equal(gcpVpcPeering.Namespace))
		})

		By("When KCP VpcPeering is Ready", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(),
					infra.KCP().Client(),
					kcpVpcPeering,
					WithConditions(KcpReadyCondition())).
				Should(Succeed(), "failed to update status on KCP VpcPeering")
		})

		By("Then SKR GcpVpcPeering is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					gcpVpcPeering,
					NewObjActions(),
					HavingConditionTrue(cloudcontrolv1beta1.ConditionTypeReady)).
				Should(Succeed(), "SKR GcpVpcPeering should be Ready, but it is not")
		})

		By("When SKR GcpVpcPeering is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), gcpVpcPeering).
				Should(Succeed(), "failed to delete SKR GcpVpcPeering")
		})

		By("Then KCP remote Network does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), remoteNetwork, WithName(gcpVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP remote Network")
		})

		By("And Then KCP VpcPeering does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcPeering, WithName(gcpVpcPeering.Status.Id)).
				Should(Succeed(), "failed to delete KCP VpcPeering")
		})

	})

})
