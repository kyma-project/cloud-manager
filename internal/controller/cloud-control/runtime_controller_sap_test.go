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

	It("Scenario: SAP Runtime with Gardener network is created and deleted", func() {

		name := "c3e935cb-d21f-45f1-890e-c5780a4557f5"
		shootName := "t-c5780a4557f5"
		secretBindingName := "secret-binding-c5780a4557f5"

		var runtime *infrastructuremanagerv1.Runtime
		subscription := &cloudcontrolv1beta1.Subscription{}
		vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

		gardenerVpcName := common.GardenerVpcName(infra.GardenerNamespace(), shootName)

		sapVpcNetworkId := "vpc-c5780a4557f5"
		sapRouterId := "router-c5780a4557f5"

		kcpsubscription.Ignore.AddName(secretBindingName)
		kcpvpcnetwork.Ignore.AddName(name)

		sapMock := infra.SapMock().NewProject()
		defer sapMock.Delete()

		By("When Runtime is created", func() {
			runtime = infrastructuremanagerv1.NewSimpleRuntimeBuilder().
				WithName(name).
				WithProvider(cloudcontrolv1beta1.ProviderOpenStack).
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
					sub.Status.Provider = cloudcontrolv1beta1.ProviderOpenStack
					sub.Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
						OpenStack: &cloudcontrolv1beta1.SubscriptionInfoOpenStack{
							DomainName: sapMock.DomainName(),
							TenantName: sapMock.ProjectName(),
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
					net.Status.Identifiers.Vpc = sapVpcNetworkId
					net.Status.Identifiers.Router = sapRouterId
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

	It("Scenario: SAP Runtime with Kyma network is created and deleted", func() {
		name := "405904d6-7bb2-4ee9-9a7f-d7e5fce84e2b"
		shootName := "t-d7e5fce84e2b"
		secretBindingName := "secret-binding-d7e5fce84e2b"
		vpcNetworkName := "bb24d71a-da88-4909-883d-386893fc3fb3"

		var runtime *infrastructuremanagerv1.Runtime
		subscription := &cloudcontrolv1beta1.Subscription{}
		vpcNetwork := &cloudcontrolv1beta1.VpcNetwork{}

		sapVpcNetworkId := "vpc-d7e5fce84e2b"
		sapRouterId := "router-d7e5fce84e2b"

		kcpsubscription.Ignore.AddName(secretBindingName)
		kcpvpcnetwork.Ignore.AddName(vpcNetworkName)

		sapMock := infra.SapMock().NewProject()
		defer sapMock.Delete()

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
					sub.Status.Provider = cloudcontrolv1beta1.ProviderOpenStack
					sub.Status.SubscriptionInfo = &cloudcontrolv1beta1.SubscriptionInfo{
						OpenStack: &cloudcontrolv1beta1.SubscriptionInfoOpenStack{
							DomainName: sapMock.DomainName(),
							TenantName: sapMock.ProjectName(),
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
				WithRegion(sapMock.RegionName()).
				Build()

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), vpcNetwork)).
				To(Succeed())

			err := composed.NewStatusPatcher(vpcNetwork).
				MutateStatus(func(n *cloudcontrolv1beta1.VpcNetwork) {
					n.SetStatusProvisioned()
					n.Status.Identifiers.Name = common.KymaVpcName(vpcNetwork.Name)
					n.Status.Identifiers.Vpc = sapVpcNetworkId
					n.Status.Identifiers.Router = sapRouterId
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
