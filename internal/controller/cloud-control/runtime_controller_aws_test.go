package cloudcontrol

import (
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	kcpvpcnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/vpcnetwork"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/utils/ptr"
)

var _ = Describe("Feature: Runtime", func() {

	It("Scenario: AWS Runtime with Gardener network is created and deleted", func() {

		name := "75c7cc86-443e-444e-aa0b-b65e8296a8a0"
		shootName := "t-0e49a90cafdc"
		secretBindingName := "secret-binding-0e49a90cafdc"

		var runtime *infrastructuremanagerv1.Runtime
		subscription := &cloudcontrolv1beta1.Subscription{}
		vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

		awsVpcNetworkId := "vpc-0e49a90cafdc"
		awsInternetGatewayId := "ig-0e49a90cafdc"

		gardenerVpcName := common.GardenerVpcName(infra.GardenerNamespace(), shootName)

		kcpsubscription.Ignore.AddName(secretBindingName)
		kcpvpcnetwork.Ignore.AddName(name)

		awsMock := infra.AwsMock().NewAccount()
		defer awsMock.Delete()

		By("When Runtime is created", func() {
			runtime = infrastructuremanagerv1.NewSimpleRuntimeBuilder().
				WithName(name).
				WithProvider(cloudcontrolv1beta1.ProviderAws).
				WithShootName(shootName).
				WithBindingName(secretBindingName).
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), runtime)).
				To(Succeed())

			_, err := composed.PatchObjAddFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, runtime, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then Subscription is created with runtime's secret binding", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions(WithName(secretBindingName))).
				Should(Succeed())
		})

		By("And Then Subscription has labels as Runtime", func() {
			for _, labelName := range cloudcontrolv1beta1.ScopeLabels {
				rVal, ok := runtime.Labels[labelName]
				Expect(ok).To(BeTrue(), "unexpected logical error - runtime should have label %s set", labelName)
				Expect(subscription.Labels).To(HaveKeyWithValue(labelName, rVal), "subscription should have label %s", labelName)
			}
		})

		By("And Then Subscription has label managed-by cloud-manager", func() {
			Expect(subscription.Labels).To(HaveKeyWithValue(util.WellKnownK8sLabelManagedBy, util.DefaultCloudManagerManagedByLabelValue))
		})

		By("And Then VpcNetwork is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork, NewObjActions(WithName(name))).
				Should(Succeed())
		})

		By("And Then VpcNetwork type is Gardener", func() {
			Expect(vpcNetwork.Spec.Type).To(Equal(cloudcontrolv1beta1.VpcNetworkTypeGardener))
		})

		By("And Then VpcNetwork is in the Subscription", func() {
			Expect(vpcNetwork.Spec.Subscription).To(Equal(subscription.Name))
		})

		By("And Then VpcNetwork name matches Gardener naming", func() {
			Expect(vpcNetwork.Spec.VpcNetworkName).To(HaveValue(Equal(gardenerVpcName)))
		})

		By("And Then VpcNetwork has labels as Runtime", func() {
			for _, labelName := range cloudcontrolv1beta1.ScopeLabels {
				rVal, ok := runtime.Labels[labelName]
				Expect(ok).To(BeTrue(), "unexpected logical error - runtime should have label %s set", labelName)
				Expect(vpcNetwork.Labels).To(HaveKeyWithValue(labelName, rVal), "vpcNetwork should have label %s", labelName)
			}
		})

		By("And Then VpcNetwork has label managed-by cloud-manager", func() {
			Expect(vpcNetwork.Labels).To(HaveKeyWithValue(util.WellKnownK8sLabelManagedBy, util.DefaultCloudManagerManagedByLabelValue))
		})

		By("When Subscription is ready", func() {
			err := composed.NewStatusPatcher(subscription).
				MutateStatus(func(sub *cloudcontrolv1beta1.Subscription) {
					sub.SetStatusReady()
					sub.Status.Provider = cloudcontrolv1beta1.ProviderAws
					sub.Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
						Aws: &cloudcontrolv1beta1.SubscriptionInfoAws{
							Account: awsMock.AccountId(),
						},
					}
				}).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("And When VpcNetwork is ready", func() {
			err := composed.NewStatusPatcher(vpcNetwork).
				MutateStatus(func(net *cloudcontrolv1beta1.VpcNetwork) {
					net.SetStatusProvisioned()
					net.Status.Identifiers.Name = ptr.Deref(vpcNetwork.Spec.VpcNetworkName, "")
					net.Status.Identifiers.Vpc = awsVpcNetworkId
					net.Status.Identifiers.InternetGateway = awsInternetGatewayId
				}).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then Runtime is annotated as handled", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), runtime, NewObjActions(), HavingAnnotation(cloudcontrolv1beta1.AnnotationRuntimeHandled, "true")).
				Should(Succeed())
		})

		// DELETE ===============================================================

		By("When Runtime is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), runtime)).To(Succeed())
		})

		By("Then VpcNetwork is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), vpcNetwork).
				Should(Succeed())
		})

		By("And Then Subscription is not deleted", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions())).
				To(Succeed())
			Expect(subscription.DeletionTimestamp).To(BeNil(), "subscription should not be marked for deletion")
		})

		By("// cleanup: delete Subscription", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), subscription)).
				To(Succeed())
			Expect(IsDeleted(infra.Ctx(), infra.KCP().Client(), subscription)).
				To(Succeed())
		})

		By("// cleanup: delete Runtime", func() {
			_, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, runtime, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
			Expect(IsDeleted(infra.Ctx(), infra.KCP().Client(), runtime)).
				To(Succeed())
		})
	})

	It("Scenario: AWS Runtime with Kyma network is created and deleted", func() {
		name := "39445a59-5aa0-4e7e-8c0b-1dd58d4b474b"
		shootName := "t-cf8c4adaa376"
		secretBindingName := "secret-binding-cf8c4adaa376"
		vpcNetworkName := "9f04a7d7-58ba-480a-9f01-95d52d9499f6"

		var runtime *infrastructuremanagerv1.Runtime
		subscription := &cloudcontrolv1beta1.Subscription{}
		vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

		awsVpcNetworkId := "vpc-cf8c4adaa376"
		awsInternetGatewayId := "ig-cf8c4adaa376"

		kcpsubscription.Ignore.AddName(secretBindingName)
		kcpvpcnetwork.Ignore.AddName(vpcNetworkName)

		awsMock := infra.AwsMock().NewAccount()
		defer awsMock.Delete()

		By("Given Subscription is created", func() {
			subscription = cloudcontrolv1beta1.NewSubscriptionBuilder().
				WithName(secretBindingName).
				WithFinalizer(api.CommonFinalizerDeletionHook).
				WithLabel(cloudcontrolv1beta1.SubscriptionLabelBindingName, secretBindingName).
				WithBindingName(secretBindingName).
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), subscription)).
				To(Succeed())

			err := composed.NewStatusPatcher(subscription).
				MutateStatus(func(sub *cloudcontrolv1beta1.Subscription) {
					sub.SetStatusReady()
					sub.Status.Provider = cloudcontrolv1beta1.ProviderAws
					sub.Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
						Aws: &cloudcontrolv1beta1.SubscriptionInfoAws{
							Account: awsMock.AccountId(),
						},
					}
				}).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Given VpcNetwork is created", func() {
			vpcNetwork = cloudcontrolv1beta1.NewVpcNetworkBuilder().
				WithName(vpcNetworkName).
				WithFinalizer(api.CommonFinalizerDeletionHook).
				WithType(cloudcontrolv1beta1.VpcNetworkTypeKyma).
				WithCidrBlocks("10.250.0.0/16").
				WithSubscription(subscription.Name).
				WithRegion("eu-west-1").
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())

			err := composed.NewStatusPatcher(vpcNetwork).
				MutateStatus(func(n *cloudcontrolv1beta1.VpcNetwork) {
					n.SetStatusProvisioned()
					n.Status.Identifiers.Name = common.KymaVpcName(vpcNetwork.Name)
					n.Status.Identifiers.Vpc = awsVpcNetworkId
					n.Status.Identifiers.InternetGateway = awsInternetGatewayId
				}).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("When Runtime is created", func() {
			runtime = infrastructuremanagerv1.NewSimpleRuntimeBuilder().
				WithName(name).
				WithShootName(shootName).
				WithBindingName(secretBindingName).
				WithVpcNetworkName(new(vpcNetworkName)).
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), runtime)).
				To(Succeed())

			_, err := composed.PatchObjAddFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, runtime, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then Runtime is annotated as handled", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), runtime, NewObjActions(), HavingAnnotation(cloudcontrolv1beta1.AnnotationRuntimeHandled, "true")).
				Should(Succeed())
		})

	})
})
