package cloudcontrol

import (
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/common"
	kcpnetwork "github.com/kyma-project/cloud-manager/pkg/kcp/network"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

var _ = Describe("Feature: KCP Scope Azure", func() {

	It("Scenario: KCP Azure Scope is created when module is activated in Kyma CR", func() {
		const (
			kymaName = "ca5b791b-87df-40ed-bea8-f10b84c483dd"
		)

		kymaNetworkName := common.KcpNetworkKymaCommonName(kymaName)
		kcpnetwork.Ignore.AddName(kymaNetworkName)

		shoot := &gardenerTypes.Shoot{}

		By("Given Shoot exists", func() {
			Eventually(CreateShootAzure).
				WithArguments(infra.Ctx(), infra, shoot, WithName(kymaName)).
				Should(Succeed(), "failed creating garden shoot for Azure")
		})

		kymaCR := util.NewKymaUnstructured()
		By("And Given Kyma CR exists", func() {
			Eventually(CreateKymaCR).
				WithArguments(infra.Ctx(), infra, kymaCR, WithName(kymaName), WithKymaSpecChannel("dev")).
				Should(Succeed(), "failed creating kyma cr")
		})

		scope := &cloudcontrolv1beta1.Scope{}

		By("Then Scope does not exist", func() {
			Consistently(LoadAndCheck, time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(
					WithName(kymaName),
					WithNamespace(DefaultKcpNamespace),
				)).
				Should(Not(Succeed()), "expected Scope not to exist")
		})

		By("And Then SKR is not active", func() {
			Expect(infra.ActiveSkrCollection().Contains(kymaName)).
				To(BeFalse(), "expected SKR not to be active, but it is active")
		})

		By("When Kyma Module state is Processing", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaCR, WithKymaStatusModuleState(util.KymaModuleStateProcessing)).
				Should(Succeed(), "failed updating KymaCR module to Processing state")
		})

		By("Then Scope is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(WithName(kymaName))).
				Should(Succeed(), "expected Scope to be created")
		})

		By("And Then Scope has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(), HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady)).
				Should(Succeed(), "expected created Scope to have Ready condition")
		})

		By("And Then Scope provider is Azure", func() {
			Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderAzure))
		})

		By("And Then Scope has spec.kymaName to equal shoot.name", func() {
			Expect(scope.Spec.KymaName).To(Equal(shoot.Name), "expected Scope.spec.kymaName to equal shoot.name")
		})

		By("And Then Scope has spec.region equal to shoot.spec.region", func() {
			Expect(scope.Spec.Region).To(Equal(shoot.Spec.Region), "expected Shoot.spec.region equal to shoot.spec.region")
		})

		By("And Then Scope has spec.scope.aws equal to nil", func() {
			Expect(scope.Spec.Scope.Aws).To(BeNil(), "expected Shoot.spec.scope.azure to be nil")
		})

		By("And Then Scope has spec.scope.gcp equal to nil", func() {
			Expect(scope.Spec.Scope.Gcp).To(BeNil(), "expected Shoot.spec.scope.gcp to be nil")
		})

		By("And Then Scope has Azure subscriptionId and tenantId", func() {
			Expect(scope.Spec.Scope.Azure).NotTo(BeNil())
			Expect(scope.Spec.Scope.Azure.SubscriptionId).NotTo(BeEmpty())
			Expect(scope.Spec.Scope.Azure.TenantId).NotTo(BeEmpty())
		})

		By("And Then Scope has spec.scope.azure.network.zones as shoot", func() {
			Expect(scope.Spec.Scope.Azure.Network.Zones).To(HaveLen(3))
			Expect(scope.Spec.Scope.Azure.Network.Zones[0].Name).To(Equal("2")) // as set in CreateShootAzure
			Expect(scope.Spec.Scope.Azure.Network.Zones[1].Name).To(Equal("3")) // as set in CreateShootAzure
			Expect(scope.Spec.Scope.Azure.Network.Zones[2].Name).To(Equal("1")) // as set in CreateShootAzure
		})

		By("And Then SKR is active", func() {
			Expect(infra.ActiveSkrCollection().Contains(kymaName)).
				To(BeTrue(), "expected SKR to be active, but it is not active")
		})

		By("And Then Kyma CR has finalizer", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaCR, NewObjActions()).
				Should(Succeed())
			Expect(controllerutil.ContainsFinalizer(kymaCR, cloudcontrolv1beta1.FinalizerName)).
				To(BeTrue(), "expected Kyma CR to have finalizer, but it does not")
		})

		kymaNetwork := &cloudcontrolv1beta1.Network{}
		By("And Then Kyma Network is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaNetwork, NewObjActions(WithName(kymaNetworkName))).
				Should(Succeed(), "expected Kyma Network to be created")
		})
	})

	It("Scenario: KCP Azure Scope is deleted when module is deactivated in Kyma CR", func() {
		const (
			kymaName = "1887bd00-3c68-4a28-ba8b-6ea66671c6f6"
		)

		shoot := &gardenerTypes.Shoot{}
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Shoot exists", func() {
			Eventually(CreateShootAzure).
				WithArguments(infra.Ctx(), infra, shoot, WithName(kymaName)).
				Should(Succeed(), "failed creating garden shoot for azure")
		})

		kymaCR := util.NewKymaUnstructured()
		By("And Given Kyma CR exists", func() {
			Eventually(CreateKymaCR).
				WithArguments(infra.Ctx(), infra, kymaCR, WithName(kymaName), WithKymaSpecChannel("dev")).
				Should(Succeed(), "failed creating kyma cr")
		})

		By("And Given module is active", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaCR, WithKymaStatusModuleState(util.KymaModuleStateReady)).
				Should(Succeed(), "failed updating KymaCR module state to Processing")

			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(WithName(kymaName)), HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady)).
				Should(Succeed(), "expected Scope to be created and have Ready condition")
		})

		By("When module is deactivated", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaCR, NewObjActions()).
				Should(Succeed(), "failed reloading Kyma CR")
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaCR, WithKymaStatusModuleState(util.KymaModuleStateNotPresent)).
				Should(Succeed(), "failed updating KymaCR module state to NotPresent")
		})

		By("Then Scope is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope).
				Should(Succeed(), "expected Scope to be deleted, but it still exists")
		})

		By("And Then Kyma CR does not have finalizer", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaCR, NewObjActions()).
				Should(Succeed(), "failed reloading Kyma CR")

			Expect(controllerutil.ContainsFinalizer(kymaCR, cloudcontrolv1beta1.FinalizerName)).
				To(BeFalse(), "expected Kyma CR not to have finalizer, but it still has it")
		})

	})

})
