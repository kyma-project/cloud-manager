package sim

import (
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: KCP SIM", func() {

	It("Scenario: Instance is created, module added and removed, and instance deleted", func() {
		const (
			alias         = "test"
			globalAccount = "6d0e57f0-66ed-4e60-9eb6-346e3354c5d0"
			subAccount    = "4d08684f-86f1-4e55-885f-0ae6920e65fb"
			provider      = cloudcontrolv1beta1.ProviderAws
		)

		var instanceDetails InstanceDetails
		rt := &infrastructuremanagerv1.Runtime{}
		shoot := &gardenertypes.Shoot{}
		//var gardenerCluster *infrastructuremanagerv1.GardenerCluster

		By("When instance is created", func() {
			id, err := simInstance.Keb().CreateInstance(infra.Ctx(), CreateInstanceInput{
				Alias:         alias,
				GlobalAccount: globalAccount,
				SubAccount:    subAccount,
				Provider:      provider,
				Region:        "",
			})
			Expect(err).NotTo(HaveOccurred())
			instanceDetails = id
		})

		By("Then Runtime is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), rt, NewObjActions(WithName(instanceDetails.RuntimeID), WithNamespace(e2econfig.Config.KcpNamespace))).
				Should(Succeed())
		})

		By("And Then Shoot is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.Garden().Client(), shoot, NewObjActions(WithName(instanceDetails.ShootName), WithNamespace(e2econfig.Config.GardenNamespace))).
				Should(Succeed())
		})

	})
})
