package e2e

import (
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	e2econfig "github.com/kyma-project/cloud-manager/e2e/config"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Feature: Scenario Session", func() {

	It("Scenario: Add existing runtime", func() {

		alias := "91e3ca4e-2204-4c57-a582-118c6b26840e"

		var instanceDetails sim.InstanceDetails
		rt := &infrastructuremanagerv1.Runtime{}
		shoot := &gardenertypes.Shoot{}

		By("Given an SKR instance exists", func() {
			id, err := world.Sim().Keb().CreateInstance(infra.Ctx(), sim.CreateInstanceInput{
				Alias:         alias,
				GlobalAccount: "5e9123a1-6c0d-4e2d-a058-24b5e6629b2e",
				SubAccount:    "b5921ea0-9283-451b-b407-4f940fb7ecf2",
				Provider:      "gcp",
			})
			Expect(err).NotTo(HaveOccurred())
			instanceDetails = id

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.Garden().Client(), shoot,
					NewObjActions(WithName(instanceDetails.ShootName), WithNamespace(e2econfig.Config.GardenNamespace)),
					sim.HavingShootReady,
				).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), rt,
					NewObjActions(WithName(instanceDetails.RuntimeID)),
					sim.HavingRuntimeProvisioningCompleted,
				).
				Should(Succeed())
		})

		session := NewScenarioSession()
		var clusterInSession ClusterInSession

		By("When Scenario session AddExistingCluster is called for existing SKR instance", func() {
			cis, err := session.AddExistingCluster(infra.Ctx(), alias)
			Expect(err).NotTo(HaveOccurred())
			clusterInSession = cis
		})

		By("Then cluster has IsCreatedInSession() == false", func() {
			Expect(clusterInSession.IsCreatedInSession()).To(BeFalse())
		})

		By("And Then cluster has IsCurrent() == true", func() {
			Expect(clusterInSession.IsCurrent()).To(BeTrue())
		})

		By("And Then cluster has RuntimeID() equal to created runtime id", func() {
			Expect(clusterInSession.RuntimeID()).To(Equal(instanceDetails.RuntimeID))
		})

		By("And Then session has current cluster", func() {
			c := session.CurrentCluster()
			Expect(c).NotTo(BeNil())
			Expect(c.Alias()).To(Equal(alias))
		})

		By("When session is terminated", func() {
			err := session.Terminate(infra.Ctx())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then SKR instance still exists", func() {
			arr, err := world.Sim().Keb().List(infra.Ctx(), sim.WithAlias(alias))
			Expect(err).NotTo(HaveOccurred())
			Expect(arr).To(HaveLen(1))
			Expect(arr[0].Alias).To(Equal(alias))
		})
	})

	It("Scenario: Add new runtime", func() {

		alias := "d5c23067-6d8c-4538-a52a-f76dda4cdafb"

		rt := &infrastructuremanagerv1.Runtime{}
		shoot := &gardenertypes.Shoot{}

		session := NewScenarioSession()
		var clusterInSession ClusterInSession

		By("When Scenario session CreateNewSkrCluster() is called", func() {
			cis, err := session.CreateNewSkrCluster(infra.Ctx(), sim.CreateInstanceInput{
				Alias:         alias,
				GlobalAccount: "038bad72-d2a3-4659-9d9c-49e2de840162",
				SubAccount:    "064ed161-6034-4905-a243-d994c392b683",
				Provider:      "gcp",
			})
			Expect(err).NotTo(HaveOccurred())
			clusterInSession = cis
		})

		By("Then SKR instance exists", func() {
			id, err := world.Sim().Keb().GetInstance(infra.Ctx(), clusterInSession.RuntimeID())
			Expect(err).NotTo(HaveOccurred())
			Expect(id).NotTo(BeNil())
		})

		By("And Then cluster has IsCreatedInSession() == true", func() {
			Expect(clusterInSession.IsCreatedInSession()).To(BeTrue())
		})

		By("And Then cluster has IsCurrent() == true", func() {
			Expect(clusterInSession.IsCurrent()).To(BeTrue())
		})

		By("And Then cluster has RuntimeID() equal to created runtime id", func() {
			Expect(clusterInSession.RuntimeID()).To(Equal(clusterInSession.RuntimeID()))
		})

		By("Then Runtime is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), rt, NewObjActions(WithName(clusterInSession.RuntimeID()))).
				Should(Succeed())
		})

		By("And Then Shoot is ready", func() {
			Expect(LoadAndCheck(
				infra.Ctx(), infra.Garden().Client(), shoot,
				NewObjActions(WithName(clusterInSession.ShootName()), WithNamespace(e2econfig.Config.GardenNamespace)),
				sim.HavingShootReady,
			)).
				To(Succeed())
		})

		By("Then Runtime has ProvisioningCompleted", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), rt, NewObjActions(), sim.HavingRuntimeProvisioningCompleted)).
				To(Succeed())
		})

		By("And Then session has current cluster", func() {
			c := session.CurrentCluster()
			Expect(c).NotTo(BeNil())
			Expect(c.Alias()).To(Equal(alias))
		})

		By("When session is terminated", func() {
			err := session.Terminate(infra.Ctx())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then eventually Runtime does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), rt).
				Should(Succeed())
		})

		By("And Then SKR instance does not exists", func() {
			arr, err := world.Sim().Keb().List(infra.Ctx(), sim.WithAlias(alias))
			Expect(err).NotTo(HaveOccurred())
			Expect(arr).To(HaveLen(0))
		})

		By("And Then Shoot does not exist", func() {
			Expect(IsDeleted(infra.Ctx(), infra.Garden().Client(), shoot)).
				To(Succeed())
		})
	})
})
