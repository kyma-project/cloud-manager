package cloudcontrol

import (
	gardenerTypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"time"
)

var _ = Describe("KCP Scope", func() {

	const (
		kymaName = "51485d74-0e28-44f9-ae80-3088128d8747"
	)
	// Set the path to an arbitrary file path to prevent errors
	os.Setenv("GCP_SA_JSON_KEY_PATH", "testdata/serviceaccount.json")
	It("Scope GCP", func() {
		shoot := &gardenerTypes.Shoot{}
		By("Given Shoot exists", func() {
			Eventually(CreateShootGcp). // for gcp, we don't really read anything of an importance from the shoot, so we can use aws
							WithArguments(infra.Ctx(), infra, shoot, WithName(kymaName)).
							Should(Succeed(), "failed creating garden shoot for gcp")
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

		By("And has provider gcp", func() {
			Expect(scope.Spec.Provider).To(Equal(cloudcontrolv1beta1.ProviderGCP))
		})

		By("And has spec.kymaName to equal shoot.name", func() {
			Expect(scope.Spec.KymaName).To(Equal(shoot.Name), "expected Scope.spec.kymaName to equal shoot.name")
		})

		By("And has nil spec.scope.azure", func() {
			Expect(scope.Spec.Scope.Azure).To(BeNil(), "expected Shoot.spec.scope.azure to be nil")
		})

		By("And has nil spec.scope.aws", func() {
			Expect(scope.Spec.Scope.Aws).To(BeNil(), "expected Shoot.spec.scope.aws to be nil")
		})
		By("And has Ready condition", func() {
			Expect(scope.Status.Conditions).To(HaveLen(1))
			Expect(scope.Status.Conditions[0].Type).To(Equal(cloudcontrolv1beta1.ConditionTypeReady))
		})
	})

})
