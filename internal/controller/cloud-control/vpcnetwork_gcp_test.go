package cloudcontrol

import (
	"cloud.google.com/go/compute/apiv1/computepb"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
)

var _ = Describe("Feature: VpcNetwork", func() {

	It("Scenario: VpcNetwork GCP is created and deleted", func() {

		const subscriptionName = "0d46c1e5-0f3f-44df-bf8e-464c5d7a6f0a"
		const vpcNetworkName = "a0dbd00b-8a9b-4018-a657-4f1f569dee96"

		const region = "us-east1"

		subscription := &cloudcontrolv1beta1.Subscription{}
		var vpcNetwork *cloudcontrolv1beta1.VpcNetwork

		gcpMock := infra.GcpMock2().NewSubscription("vpcnetwork-gcp")

		By("Given GCP Subscription exists", func() {
			kcpsubscription.Ignore.AddName(subscriptionName)
			Expect(
				CreateSubscription(infra.Ctx(), infra, subscription,
					WithName(subscriptionName),
					WithSubscriptionSpecGarden("binding-name")),
			).To(Succeed())

			Expect(
				SubscriptionPatchStatusReadyGcp(infra.Ctx(), infra, subscription, gcpMock.ProjectId()),
			).To(Succeed())
		})

		By("When VpcNetwork is created", func() {
			vpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithRegion(region).
				WithSubscription(subscriptionName).
				WithCidrBlocks("10.250.0.0/16").
				Build()

			Expect(
				CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork, WithName(vpcNetworkName)),
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
			Expect(vpcNetwork.Labels[cloudcontrolv1beta1.SubscriptionLabel]).To(Equal(subscriptionName))
		})

		By("Then VpcNetwork has provider label", func() {
			Expect(vpcNetwork.Labels[cloudcontrolv1beta1.SubscriptionLabelProvider]).To(Equal(string(cloudcontrolv1beta1.ProviderGCP)))
		})

		By("Then VpcNetwork status has vpcID", func() {
			Expect(vpcNetwork.Status.Identifiers.Vpc).NotTo(BeEmpty())
		})

		By("Then VpcNetwork status has routerID", func() {
			Expect(vpcNetwork.Status.Identifiers.Router).NotTo(BeEmpty())
		})

		var gcpNetwork *computepb.Network
		var gcpRouter *computepb.Router

		By("Then GPC VPC Network exists", func() {
			aVpc, err := gcpMock.GetNetwork(infra.Ctx(), &computepb.GetNetworkRequest{
				Project: gcpMock.ProjectId(),
				Network: vpcNetwork.Status.Identifiers.Vpc,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(aVpc).ToNot(BeNil())
			gcpNetwork = aVpc
		})

		By("Then GPC Cloud Router exists", func() {
			aRouter, err := gcpMock.GetRouter(infra.Ctx(), &computepb.GetRouterRequest{
				Project: gcpMock.ProjectId(),
				Region:  region,
				Router:  vpcNetwork.Status.Identifiers.Router,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(aRouter).ToNot(BeNil())
			gcpRouter = aRouter
		})

		// DELETE =================================================

		By("When KCP VpcNetwork is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())
		})

		By("Then KCP VpcNetwork does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork).
				Should(Succeed())
		})

		By("Then GCP VPC Network does not exist", func() {
			aVpc, err := gcpMock.GetNetwork(infra.Ctx(), &computepb.GetNetworkRequest{
				Project: gcpMock.ProjectId(),
				Network: gcpNetwork.GetName(),
			})
			Expect(err).To(HaveOccurred())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
			Expect(aVpc).To(BeNil())
		})

		By("Then GCP Cloud Router does not exist", func() {
			aRouter, err := gcpMock.GetRouter(infra.Ctx(), &computepb.GetRouterRequest{
				Project: gcpMock.ProjectId(),
				Region:  gcpRouter.GetRegion(),
				Router:  gcpRouter.GetName(),
			})
			Expect(err).To(HaveOccurred())
			Expect(gcpmeta.IsNotFound(err)).To(BeTrue())
			Expect(aRouter).To(BeNil())
		})

		By("// cleanup: delete Subscription", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), subscription)).To(Succeed())
		})
	})

})
