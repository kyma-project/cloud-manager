package cloudcontrol

import (
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Scope AWS", Focus, Ordered, func() {

	const (
		kymaName = "5d60be8c-e422-48ff-bd0a-166b0e09dc58"
	)

	shoot := &gardenerTypes.Shoot{}

	It("Given Shoot exists", func() {
		Eventually(CreateShootAws).
			WithArguments(infra.Ctx(), infra, shoot, WithName(kymaName)).
			Should(Succeed(), "failed creating garden shoot for aws")
	})

	kymaCR := util.NewKymaUnstructured()

	It("And Given Kyma CR exists", func() {
		Eventually(CreateKymaCR).
			WithArguments(infra.Ctx(), infra, kymaCR, WithName(kymaName)).
			Should(Succeed(), "failed creating kyma cr")
	})

	It("Then Scope should not exist", func() {
		scope := &cloudcontrolv1beta1.Scope{}
		Consistently(LoadAndCheck).
			WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(
				WithName(kymaName),
				WithNamespace(DefaultKcpNamespace),
			)).
			Should(Not(Succeed()), "expected Scope not to exist")
	})

	It("When Kyma Module state is Ready", func() {
		Eventually(KymaCRModuleStateUpdate).
			WithArguments(infra.Ctx(), infra.KCP().Client(), kymaCR, WithKymaModuleState(util.KymaModuleStateReady)).
			Should(Succeed(), "failed updating KymaCR module to Ready state")
	})

	scope := &cloudcontrolv1beta1.Scope{}

	It("Then Scope is created", func() {
		Eventually(LoadAndCheck).
			WithArguments(infra.Ctx(), infra.KCP().Client(), scope, NewObjActions(WithName(kymaName))).
			Should(Succeed(), "expected Scope to be created")

		By("And has provider aws")
		Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderAws))

		By("And has spec.kymaName to equal shoot.name")
		Expect(scope.Spec.KymaName).To(Equal(shoot.Name), "expected Scope.spec.kymaName to equal shoot.name")

		By("And has spec.region equal to shoot.spec.region")
		Expect(scope.Spec.Region).To(Equal(shoot.Spec.Region), "expected Shoot.spec.region equal to shoot.spec.region")

		By("And has nil spec.scope.azure")
		Expect(scope.Spec.Scope.Azure).To(BeNil(), "expected Shoot.spec.scope.azure to be nil")

		By("And has nil spec.scope.gcp")
		Expect(scope.Spec.Scope.Gcp).To(BeNil(), "expected Shoot.spec.scope.gcp to be nil")

		By("And has spec.scope.aws.accountId")
		Expect(scope.Spec.Scope.Aws).NotTo(BeNil())
		Expect(scope.Spec.Scope.Aws.AccountId).NotTo(BeEmpty())
		Expect(scope.Spec.Scope.Aws.AccountId).To(Equal(infra.AwsMock().GetAccount()))

		By("And has spec.scope.aws.network.zones as shoot")
		Expect(scope.Spec.Scope.Aws.Network.Zones).To(HaveLen(3))
		Expect(scope.Spec.Scope.Aws.Network.Zones[0].Name).To(Equal("eu-west-1a")) // as set in GivenGardenShootAwsExists
		Expect(scope.Spec.Scope.Aws.Network.Zones[1].Name).To(Equal("eu-west-1b")) // as set in GivenGardenShootAwsExists
		Expect(scope.Spec.Scope.Aws.Network.Zones[2].Name).To(Equal("eu-west-1c")) // as set in GivenGardenShootAwsExists
	})

})
