package cloudcontrol

import (
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

var _ = Describe("Feature: KCP Scope", func() {

	const (
		kymaName = "5d60be8c-e422-48ff-bd0a-166b0e09dc58"
	)

	It("Scenario: KCP AWS Scope is created when module is activated in Kyma CR", func() {
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
		By("Then Scope should not exist", func() {
			Consistently(LoadAndCheck, time.Second).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(
					WithName(kymaName),
					WithNamespace(DefaultKcpNamespace),
				)).
				Should(Not(Succeed()), "expected Scope not to exist")
		})

		By("When Kyma Module is listed in status", func() {
			Eventually(UpdateStatus).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kymaCR, WithKymaStatusModuleState(util.KymaModuleStateProcessing)).
				Should(Succeed(), "failed updating KymaCR module to Processing state")
		})

		By("Then Scope is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(WithName(kymaName))).
				Should(Succeed(), "expected Scope to be created")
		})

		By("And has provider aws", func() {
			Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderAws))
		})

		By("And has spec.kymaName to equal shoot.name", func() {
			Expect(scope.Spec.KymaName).To(Equal(shoot.Name), "expected Scope.spec.kymaName to equal shoot.name")
		})

		By("And has spec.region equal to shoot.spec.region", func() {
			Expect(scope.Spec.Region).To(Equal(shoot.Spec.Region), "expected Shoot.spec.region equal to shoot.spec.region")
		})

		By("And has nil spec.scope.azure", func() {
			Expect(scope.Spec.Scope.Azure).To(BeNil(), "expected Shoot.spec.scope.azure to be nil")
		})

		By("And has nil spec.scope.gcp", func() {
			Expect(scope.Spec.Scope.Gcp).To(BeNil(), "expected Shoot.spec.scope.gcp to be nil")
		})

		By("And has spec.scope.aws.accountId", func() {
			Expect(scope.Spec.Scope.Aws).NotTo(BeNil())
			Expect(scope.Spec.Scope.Aws.AccountId).NotTo(BeEmpty())
			Expect(scope.Spec.Scope.Aws.AccountId).To(Equal(infra.AwsMock().GetAccount()))
		})

		By("And has spec.scope.aws.network.zones as shoot", func() {
			Expect(scope.Spec.Scope.Aws.Network.Zones).To(HaveLen(3))
			Expect(scope.Spec.Scope.Aws.Network.Zones[0].Name).To(Equal("eu-west-1a")) // as set in GivenGardenShootAwsExists
			Expect(scope.Spec.Scope.Aws.Network.Zones[1].Name).To(Equal("eu-west-1b")) // as set in GivenGardenShootAwsExists
			Expect(scope.Spec.Scope.Aws.Network.Zones[2].Name).To(Equal("eu-west-1c")) // as set in GivenGardenShootAwsExists
		})
	})

})
