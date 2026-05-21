package cloudcontrol

import (
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/elliotchance/pie/v2"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: VpcNetwork", func() {

	It("Scenario: AWS VpcNetwork type Kyma is created and deleted", func() {
		const subscriptionName = "dd48fd32-7ae9-4fe3-aa24-d66cb1ea06df"
		const vpcNetworkName = "3262da42-3fa7-485f-9487-bc66a5fcacc2"
		const region = "us-east-1"

		subscription := &cloudcontrolv1beta1.Subscription{}
		var vpcNetwork *cloudcontrolv1beta1.VpcNetwork

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()

		By("Given AWS Subscription exists", func() {
			kcpsubscription.Ignore.AddName(subscriptionName)
			Expect(
				CreateSubscription(infra.Ctx(), infra, subscription,
					WithName(subscriptionName),
					WithSubscriptionSpecGarden("binding-name")),
			).To(Succeed())

			Expect(
				SubscriptionPatchStatusReadyAws(infra.Ctx(), infra, subscription, awsAccount.AccountId()),
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
			Expect(vpcNetwork.Labels[cloudcontrolv1beta1.SubscriptionLabelProvider]).To(Equal(string(cloudcontrolv1beta1.ProviderAws)))
		})

		By("Then VpcNetwork status has vpcID", func() {
			Expect(vpcNetwork.Status.Identifiers.Vpc).NotTo(BeEmpty())
		})

		By("Then VpcNetwork status has name", func() {
			Expect(vpcNetwork.Status.Identifiers.Name).NotTo(BeEmpty())
		})

		awsMock := awsAccount.Region(region)

		var vpc *ec2types.Vpc

		By("Then AWS VPC Network exists", func() {
			aVpc, err := awsMock.DescribeVpc(infra.Ctx(), vpcNetwork.Status.Identifiers.Vpc)
			Expect(err).ToNot(HaveOccurred())
			Expect(aVpc).ToNot(BeNil())
			vpc = aVpc
		})

		By("Then AWS VPC Network has correct CIDR block", func() {
			Expect(pie.Map(vpc.CidrBlockAssociationSet, func(x ec2types.VpcCidrBlockAssociation) string {
				return ptr.Deref(x.CidrBlock, "")
			})).To(Equal(vpcNetwork.Status.CidrBlocks))
		})

		// DELETE ===============================================

		By("When KCP VpcNetwork is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())
		})

		By("Then KCP VpcNetwork does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork).
				Should(Succeed())
		})

		By("Then AWS VPC Network does not exist", func() {
			vpc, err := awsMock.DescribeVpc(infra.Ctx(), ptr.Deref(vpc.VpcId, ""))
			Expect(err).ToNot(HaveOccurred())
			Expect(vpc).To(BeNil())
		})

		By("// cleanup: delete Subscription", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), subscription)).To(Succeed())
		})
	})

	It("Scenario: AWS VpcNetwork type Gardener is created and deleted", func() {
		const subscriptionName = "48c381e6-5c42-4b5a-b7f3-afea6b902a97"
		const bindingName = "binding-afea6b902a97"
		const shootName = "t-afea6b902a97"
		const kcpVpcNetworkName = "f64f3291-a1b0-4e4e-ae70-d53bbd9ec76b"
		gardenerNetworkName := common.GardenerVpcName(infra.GardenerNamespace(), shootName)
		const region = "us-east-1"

		kcpSubscription := &cloudcontrolv1beta1.Subscription{}

		awsAccount := infra.AwsMock().NewAccount()
		defer awsAccount.Delete()
		awsMock := awsAccount.Region(region)

		var kcpVpcNetwork *cloudcontrolv1beta1.VpcNetwork

		var awsVpcNetwork *ec2types.Vpc
		var awsInternetGateway *ec2types.InternetGateway

		By("Given Subscription exists", func() {
			kcpsubscription.Ignore.AddName(subscriptionName)
			Expect(
				CreateSubscription(infra.Ctx(), infra, kcpSubscription,
					WithName(subscriptionName),
					WithSubscriptionSpecGarden(bindingName)),
			).To(Succeed())

			Expect(
				SubscriptionPatchStatusReadyAws(infra.Ctx(), infra, kcpSubscription, awsAccount.AccountId()),
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
					HavingCondition(cloudcontrolv1beta1.ConditionTypeReady, metav1.ConditionFalse, cloudcontrolv1beta1.ReasonProviderError, "VPC not found")).
				Should(Succeed())
		})

		By("When AWS VPC Network is created", func() {
			net, err := awsMock.CreateVpc(infra.Ctx(), gardenerNetworkName, "10.250.0.0/16", nil)
			Expect(err).ToNot(HaveOccurred())
			awsVpcNetwork = net
		})

		By("Then VpcNetwork has provider error internet gateway not found condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork, NewObjActions(),
					HavingCondition(cloudcontrolv1beta1.ConditionTypeReady, metav1.ConditionFalse, cloudcontrolv1beta1.ReasonProviderError, "Internet gateway not found")).
				Should(Succeed())
		})

		By("When AWS Internet Gateway is created", func() {
			gw, err := awsMock.CreateInternetGateway(infra.Ctx(), gardenerNetworkName)
			Expect(err).ToNot(HaveOccurred())
			err = awsMock.AttachInternetGateway(infra.Ctx(), ptr.Deref(awsVpcNetwork.VpcId, ""), ptr.Deref(gw.InternetGatewayId, ""))
			Expect(err).ToNot(HaveOccurred())
			awsInternetGateway = gw
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
			Expect(kcpVpcNetwork.Status.Identifiers.Vpc).To(Equal(ptr.Deref(awsVpcNetwork.VpcId, "")))
		})

		By("Then KCP VpcNetwork has status.identifies.internetGateway", func() {
			Expect(kcpVpcNetwork.Status.Identifiers.InternetGateway).To(Equal(ptr.Deref(awsInternetGateway.InternetGatewayId, "")))
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
