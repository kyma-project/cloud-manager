package cloudcontrol

import (
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	alicloudmock "github.com/kyma-project/cloud-manager/pkg/kcp/provider/alicloud/mock"
	kcpscope "github.com/kyma-project/cloud-manager/pkg/kcp/scope"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: VpcNetwork", func() {

	It("Scenario: Alicloud VpcNetwork type Gardener observes existing VPC", func() {
		const (
			subscriptionName  = "ac-vpcnet-gardener-sub-01"
			bindingName       = "binding-ac-vpcnet-01"
			shootName         = "t-ac-vpcnet-01"
			kcpVpcNetworkName = "ac-vpcnet-gardener-01"
			region            = "ap-southeast-1"
		)

		gardenerNetworkName := common.GardenerVpcName(infra.GardenerNamespace(), shootName)
		// Alicloud Gardener VPCs are named "<shootName>-vpc"
		alicloudVpcName := gardenerNetworkName + "-vpc"

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()
		alicloudRegion := alicloudAccount.Region(region)

		kcpSubscription := &cloudcontrolv1beta1.Subscription{}

		By("Given Subscription exists", func() {
			kcpsubscription.Ignore.AddName(subscriptionName)
			Expect(
				CreateSubscription(infra.Ctx(), infra, kcpSubscription,
					WithName(subscriptionName),
					WithSubscriptionSpecGarden(bindingName)),
			).To(Succeed())
			Expect(
				SubscriptionPatchStatusReadyAlicloud(infra.Ctx(), infra, kcpSubscription, alicloudAccount.Credentials().AccessKeyId),
			).To(Succeed())
		})

		kcpScope := &cloudcontrolv1beta1.Scope{}

		By("And Given Scope exists", func() {
			kcpscope.Ignore.AddName(shootName)
			Eventually(CreateScopeAlicloud).
				WithArguments(infra.Ctx(), infra, kcpScope,
					alicloudAccount.Credentials().AccessKeyId,
					WithName(shootName)).
				Should(Succeed())
		})

		var kcpVpcNetwork *cloudcontrolv1beta1.VpcNetwork

		By("When VpcNetwork type Gardener is created", func() {
			kcpVpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithName(kcpVpcNetworkName).
				WithType(cloudcontrolv1beta1.VpcNetworkTypeGardener).
				WithVpcNetworkName(new(gardenerNetworkName)).
				WithRegion(region).
				WithSubscription(subscriptionName).
				WithCidrBlocks("10.180.0.0/16").
				Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork)).To(Succeed())
		})

		By("Then VpcNetwork has provider error vpc not found condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork, NewObjActions(),
					HavingCondition(cloudcontrolv1beta1.ConditionTypeReady, metav1.ConditionFalse, cloudcontrolv1beta1.ReasonProviderError, "VPC not found")).
				Should(Succeed())
		})

		var alicloudVpc *alicloudmock.VpcEntry

		By("When Alicloud VPC is created (with Gardener -vpc suffix)", func() {
			alicloudVpc = alicloudRegion.AddVpc("vpc-test-01", alicloudVpcName, "10.180.0.0/16")
		})

		By("Then VpcNetwork is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork, NewObjActions(),
					HavingConditionReasonTrue(cloudcontrolv1beta1.ConditionTypeReady, cloudcontrolv1beta1.ReasonProvisioned)).
				Should(Succeed())
		})

		By("And Then VpcNetwork has status.identifiers.vpc equal to Alicloud VPC id", func() {
			Expect(kcpVpcNetwork.Status.Identifiers.Vpc).To(Equal(alicloudVpc.VpcId))
		})

		By("And Then VpcNetwork has status.identifiers.name", func() {
			Expect(kcpVpcNetwork.Status.Identifiers.Name).To(Equal(gardenerNetworkName))
		})

		By("And Then VpcNetwork has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(kcpVpcNetwork, api.CommonFinalizerDeletionHook)).To(BeTrue())
		})

		By("And Then VpcNetwork has subscription label", func() {
			Expect(kcpVpcNetwork.Labels[cloudcontrolv1beta1.SubscriptionLabel]).To(Equal(subscriptionName))
		})

		// DELETE ================================================

		By("When KCP VpcNetwork is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork)).To(Succeed())
		})

		By("Then KCP VpcNetwork does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpVpcNetwork).
				Should(Succeed())
		})

		By("// cleanup: delete Subscription", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), kcpSubscription)).To(Succeed())
		})

		By("// cleanup: delete Scope", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), kcpScope)).To(Succeed())
		})
	})

	It("Scenario: Alicloud VpcNetwork type Kyma is created and deleted", func() {
		const (
			subscriptionName = "ac-vpcnet-kyma-sub-01"
			vpcNetworkName   = "ac-vpcnet-kyma-01"
			region           = "ap-southeast-1"
		)

		alicloudAccount := infra.AlicloudMock().NewAccount()
		defer alicloudAccount.Delete()
		alicloudRegion := alicloudAccount.Region(region)

		subscription := &cloudcontrolv1beta1.Subscription{}
		var vpcNetwork *cloudcontrolv1beta1.VpcNetwork

		By("Given Alicloud Subscription exists", func() {
			kcpsubscription.Ignore.AddName(subscriptionName)
			Expect(
				CreateSubscription(infra.Ctx(), infra, subscription,
					WithName(subscriptionName),
					WithSubscriptionSpecGarden("binding-name")),
			).To(Succeed())
			Expect(
				SubscriptionPatchStatusReadyAlicloud(infra.Ctx(), infra, subscription, alicloudAccount.Credentials().AccessKeyId),
			).To(Succeed())
		})

		By("When VpcNetwork is created", func() {
			vpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithRegion(region).
				WithSubscription(subscriptionName).
				WithCidrBlocks("10.180.0.0/16").
				Build()
			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork, WithName(vpcNetworkName))).To(Succeed())
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

		By("Then VpcNetwork status has vpcID", func() {
			Expect(vpcNetwork.Status.Identifiers.Vpc).NotTo(BeEmpty())
		})

		By("Then Alicloud VPC exists in mock", func() {
			result, err := alicloudRegion.IpRangeClient().DescribeVpcs(infra.Ctx(), vpcNetwork.Status.Identifiers.Name)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).NotTo(BeEmpty())
		})

		// DELETE ===============================================

		By("When KCP VpcNetwork is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).To(Succeed())
		})

		By("Then KCP VpcNetwork does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork).
				Should(Succeed())
		})

		By("// cleanup: delete Subscription", func() {
			Expect(infra.KCP().Client().Delete(infra.Ctx(), subscription)).To(Succeed())
		})
	})
})
