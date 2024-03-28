package cloudcontrol

import (
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"time"
)

var _ = Describe("Feature: KCP Scope", func() {

	It("Scenario: KCP AWS Scope is created when module is activated in Kyma CR", func() {
		const (
			kymaName = "5d60be8c-e422-48ff-bd0a-166b0e09dc58"
		)

		shoot := &gardenerTypes.Shoot{}

		By("Given Shoot exists", func() {
			Eventually(CreateShootAws).
				WithArguments(infra.Ctx(), infra, shoot, WithName(kymaName)).
				Should(Succeed(), "failed creating garden shoot for aws")
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

		By("And Then Scope provider is aws", func() {
			Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderAws))
		})

		By("And Then Scope has spec.kymaName to equal shoot.name", func() {
			Expect(scope.Spec.KymaName).To(Equal(shoot.Name), "expected Scope.spec.kymaName to equal shoot.name")
		})

		By("And Then Scope has spec.region equal to shoot.spec.region", func() {
			Expect(scope.Spec.Region).To(Equal(shoot.Spec.Region), "expected Shoot.spec.region equal to shoot.spec.region")
		})

		By("And Then Scope has spec.scope.azure equal to nil", func() {
			Expect(scope.Spec.Scope.Azure).To(BeNil(), "expected Shoot.spec.scope.azure to be nil")
		})

		By("And Then Scope has spec.scope.gcp equal to nil", func() {
			Expect(scope.Spec.Scope.Gcp).To(BeNil(), "expected Shoot.spec.scope.gcp to be nil")
		})

		By("And Then Scope has spec.scope.aws.accountId", func() {
			Expect(scope.Spec.Scope.Aws).NotTo(BeNil())
			Expect(scope.Spec.Scope.Aws.AccountId).NotTo(BeEmpty())
			Expect(scope.Spec.Scope.Aws.AccountId).To(Equal(infra.AwsMock().GetAccount()))
		})

		By("And Then Scope has spec.scope.aws.network.zones as shoot", func() {
			Expect(scope.Spec.Scope.Aws.Network.Zones).To(HaveLen(3))
			Expect(scope.Spec.Scope.Aws.Network.Zones[0].Name).To(Equal("eu-west-1a")) // as set in GivenGardenShootAwsExists
			Expect(scope.Spec.Scope.Aws.Network.Zones[1].Name).To(Equal("eu-west-1b")) // as set in GivenGardenShootAwsExists
			Expect(scope.Spec.Scope.Aws.Network.Zones[2].Name).To(Equal("eu-west-1c")) // as set in GivenGardenShootAwsExists
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
	})

	It("Scenario: KCP AWS Scope is deleted when module is deactivated in Kyma CR", func() {
		const (
			kymaName = "d55ac1aa-288c-4af4-a0b7-96ce5b81046b"
		)

		shoot := &gardenerTypes.Shoot{}
		scope := &cloudcontrolv1beta1.Scope{}

		By("Given Shoot exists", func() {
			Eventually(CreateShootAws).
				WithArguments(infra.Ctx(), infra, shoot, WithName(kymaName)).
				Should(Succeed(), "failed creating garden shoot for aws")
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
