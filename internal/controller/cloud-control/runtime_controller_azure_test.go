package cloudcontrol

import (
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	commongardener "github.com/kyma-project/cloud-manager/pkg/common/gardener"
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

	It("Scenario: Azure Runtime with Gardener network is created and deleted", func() {

		name := "e9ab5650-ca34-4426-913f-4b1ef0df0d4d"
		shootName := "t-4b1ef0df0d4d"
		secretBindingName := "secret-binding-4b1ef0df0d4d"

		var runtime *infrastructuremanagerv1.Runtime
		subscription := &cloudcontrolv1beta1.Subscription{}
		vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

		gardenerVpcName := common.GardenerVpcName(infra.GardenerNamespace(), shootName)

		azureResourceGroup := gardenerVpcName
		azureVpcNetworkId := "vpc-4b1ef0df0d4d"

		kcpsubscription.Ignore.AddName(secretBindingName)
		kcpvpcnetwork.Ignore.AddName(name)

		azureMock := infra.AzureMock().NewSubscription()
		defer azureMock.Delete()

		By("When Runtime is created", func() {
			runtime = infrastructuremanagerv1.NewSimpleRuntimeBuilder().
				WithName(name).
				WithProvider(cloudcontrolv1beta1.ProviderAzure).
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
					sub.Status.Provider = cloudcontrolv1beta1.ProviderAzure
					sub.Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
						Azure: &cloudcontrolv1beta1.SubscriptionInfoAzure{
							TenantId:       azureMock.TenantId(),
							SubscriptionId: azureMock.SubscriptionId(),
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
					net.Status.Identifiers.Vpc = azureVpcNetworkId
					net.Status.Identifiers.ResourceGroup = azureResourceGroup
				}).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then Runtime is annotated as security handled", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), runtime, NewObjActions(), HavingAnnotation(cloudcontrolv1beta1.RuntimeSecurityStatusAnnotation, "Ready")).
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

	It("Scenario: Azure Runtime with Kyma network is created and deleted", func() {
		name := "91b4634b-d8fb-48c2-a5c5-90719296e8d0"
		shootName := "t-90719296e8d0"
		secretBindingName := "secret-binding-90719296e8d0"
		vpcNetworkName := "1b5c64ba-8593-4bad-a242-ae78594227dd"

		var runtime *infrastructuremanagerv1.Runtime
		subscription := &cloudcontrolv1beta1.Subscription{}
		vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

		gardenerVpcName := (func() string {
			ns, err := commongardener.DefaultGardenerNamespaceProvider().GetGardenerNamespace(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
			return common.GardenerVpcName(ns, shootName)
		})()

		azureResourceGroup := gardenerVpcName
		azureVpcNetworkId := "vpc-4b1ef0df0d4d"

		kcpsubscription.Ignore.AddName(secretBindingName)
		kcpvpcnetwork.Ignore.AddName(vpcNetworkName)

		azureMock := infra.AzureMock().NewSubscription()

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
					sub.Status.Provider = cloudcontrolv1beta1.ProviderAzure
					sub.Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
						Azure: &cloudcontrolv1beta1.SubscriptionInfoAzure{
							TenantId:       azureMock.TenantId(),
							SubscriptionId: azureMock.SubscriptionId(),
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
				WithRegion("westeurope").
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())

			err := composed.NewStatusPatcher(vpcNetwork).
				MutateStatus(func(n *cloudcontrolv1beta1.VpcNetwork) {
					n.SetStatusProvisioned()
					n.Status.Identifiers.Name = common.KymaVpcName(vpcNetwork.Name)
					n.Status.Identifiers.Vpc = azureVpcNetworkId
					n.Status.Identifiers.ResourceGroup = azureResourceGroup
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

		By("Then Runtime is annotated as security handled", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), runtime, NewObjActions(), HavingAnnotation(cloudcontrolv1beta1.RuntimeSecurityStatusAnnotation, "Ready")).
				Should(Succeed())
		})

	})
})
