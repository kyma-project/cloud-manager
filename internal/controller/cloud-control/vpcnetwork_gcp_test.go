package cloudcontrol

import (
	"cloud.google.com/go/compute/apiv1/computepb"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	gcpmeta "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/meta"
	gcpvpcnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/provider/gcp/vpcnetwork"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: VpcNetwork", func() {

	It("Scenario: GCP VpcNetwork type Kyma is created and deleted", func() {

		const subscriptionName = "0d46c1e5-0f3f-44df-bf8e-464c5d7a6f0a"
		const vpcNetworkName = "a0dbd00b-8a9b-4018-a657-4f1f569dee96"

		const region = "us-east1"

		subscription := &cloudcontrolv1beta1.Subscription{}
		var vpcNetwork *cloudcontrolv1beta1.VpcNetwork

		gcpMock := infra.GcpMock2().NewSubscription("vpcnetkyma0f3f")

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

		By("Then VpcNetwork status has name", func() {
			Expect(vpcNetwork.Status.Identifiers.Name).NotTo(BeEmpty())
		})

		By("Then VpcNetwork status has vpcID", func() {
			Expect(vpcNetwork.Status.Identifiers.Vpc).NotTo(BeEmpty())
		})

		By("Then VpcNetwork status has routerID", func() {
			Expect(vpcNetwork.Status.Identifiers.Router).NotTo(BeEmpty())
		})

		var gcpNetwork *computepb.Network
		var gcpRouter *computepb.Router

		By("Then GCP VPC Network exists", func() {
			aVpc, err := gcpMock.GetNetwork(infra.Ctx(), &computepb.GetNetworkRequest{
				Project: gcpMock.ProjectId(),
				Network: vpcNetwork.Status.Identifiers.Vpc,
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(aVpc).ToNot(BeNil())
			gcpNetwork = aVpc
		})

		By("Then GCP Cloud Router exists", func() {
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

	It("Scenario: GCP VpcNetwork type Gardener is created and deleted", func() {
		const subscriptionName = "1ff23889-6061-40cf-90f0-60a52c7bf58e"
		const shootName = "t-60a52c7bf58e"
		const bindingName = "binding-60a52c7bf58e"
		const kcpVpcNetworkName = "25af538d-3da8-4752-a0f6-2931e48a80ff"
		const region = "us-east1"
		gardenerNetworkName := common.GardenerVpcName(infra.GardenerNamespace(), shootName)

		gcpMock := infra.GcpMock2().NewSubscription("vpcnetgardener6061")

		kcpSubscription := &cloudcontrolv1beta1.Subscription{}
		var kcpVpcNetwork *cloudcontrolv1beta1.VpcNetwork

		var gcpNetwork *computepb.Network
		var gcpRouter *computepb.Router

		By("Given Subscription exists", func() {
			kcpsubscription.Ignore.AddName(subscriptionName)
			Expect(
				CreateSubscription(infra.Ctx(), infra, kcpSubscription,
					WithName(subscriptionName),
					WithSubscriptionSpecGarden(bindingName)),
			).To(Succeed())

			Expect(
				SubscriptionPatchStatusReadyGcp(infra.Ctx(), infra, kcpSubscription, gcpMock.ProjectId()),
			).To(Succeed())
		})

		By("When VpcNetwork is created", func() {
			kcpVpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithName(kcpVpcNetworkName).
				WithType(cloudcontrolv1beta1.VpcNetworkTypeGardener).
				WithVpcNetworkName(new(gardenerNetworkName)).
				WithRegion(region).
				WithSubscription(subscriptionName).
				WithCidrBlocks("10.250.0.0/16").
				Build()
			err := CreateObj(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Then VpcNetwork has provider error vpc not found condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork, NewObjActions(),
					HavingCondition(cloudcontrolv1beta1.ConditionTypeReady, metav1.ConditionFalse, cloudcontrolv1beta1.ReasonProviderError, "GCP network not found")).
				Should(Succeed())
		})

		By("When GCP VPC Network is created", func() {
			op, err := gcpMock.InsertNetwork(infra.Ctx(), &computepb.InsertNetworkRequest{
				Project: gcpMock.ProjectId(),
				NetworkResource: &computepb.Network{
					Name:                  new(gardenerNetworkName),
					AutoCreateSubnetworks: new(false),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			err = op.Wait(infra.Ctx())
			Expect(err).ToNot(HaveOccurred())

			net, err := gcpMock.GetNetwork(infra.Ctx(), &computepb.GetNetworkRequest{
				Project: gcpMock.ProjectId(),
				Network: gardenerNetworkName,
			})
			Expect(err).ToNot(HaveOccurred())
			gcpNetwork = net
		})

		By("Then VpcNetwork has provider error router not found condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork, NewObjActions(),
					HavingCondition(cloudcontrolv1beta1.ConditionTypeReady, metav1.ConditionFalse, cloudcontrolv1beta1.ReasonProviderError, "GCP router not found")).
				Should(Succeed())
		})

		By("When GCP router is created", func() {
			op, err := gcpMock.InsertRouter(infra.Ctx(), &computepb.InsertRouterRequest{
				Project: gcpMock.ProjectId(),
				Region:  region,
				RouterResource: &computepb.Router{
					Name:    new(gcpvpcnetwork.RouterName(gardenerNetworkName)),
					Network: new(gcpNetwork.GetSelfLink()),
				},
			})
			Expect(err).ToNot(HaveOccurred())
			err = op.Wait(infra.Ctx())
			Expect(err).ToNot(HaveOccurred())

			r, err := gcpMock.GetRouter(infra.Ctx(), &computepb.GetRouterRequest{
				Project: gcpMock.ProjectId(),
				Region:  region,
				Router:  gcpvpcnetwork.RouterName(gardenerNetworkName),
			})
			Expect(err).ToNot(HaveOccurred())
			gcpRouter = r
		})

		By("Then VpcNetwork is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork, NewObjActions(),
					HavingConditionReasonTrue(cloudcontrolv1beta1.ConditionTypeReady, cloudcontrolv1beta1.ReasonProvisioned)).
				Should(Succeed())
		})

		By("Then KCP VpcNetwork has status.identifies.name", func() {
			Expect(kcpVpcNetwork.Status.Identifiers.Name).To(Equal(gardenerNetworkName))
		})

		By("Then KCP VpcNetwork has status.identifies.vpc", func() {
			Expect(kcpVpcNetwork.Status.Identifiers.Vpc).To(Equal(gcpNetwork.GetName()))
		})

		By("Then KCP VpcNetwork has status.identifies.router", func() {
			Expect(kcpVpcNetwork.Status.Identifiers.Router).To(Equal(gcpRouter.GetName()))
		})

		By("Then KCP VpcNetwork has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(kcpVpcNetwork, api.CommonFinalizerDeletionHook)).
				To(BeTrue())
		})

		// DELETE ================================================

		By("When KCP VpcNetwork is deleted", func() {
			err := Delete(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork)
			Expect(err).ToNot(HaveOccurred())
		})

		By("Then KCP VpcNetwork does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork).
				Should(Succeed())
		})

		By("// cleanup: delete subscription", func() {
			err := Delete(infra.Ctx(), infra.KCP().Client(), kcpSubscription)
			Expect(err).ToNot(HaveOccurred())
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpSubscription).
				Should(Succeed())
		})

	})
})
