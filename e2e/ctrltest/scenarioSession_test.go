package ctrltest

import (
	"time"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/e2e"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/e2e/sim"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
)

var _ = Describe("Feature: New and existing clusters added to Scenario Session", func() {

	It("Scenario: Add existing runtime does not delete SKR instance on session terminate", func() {

		alias := "91e3ca4e-2204-4c57-a582-118c6b26840e"

		var instanceDetails e2ekeb.InstanceDetails
		rt := &infrastructuremanagerv1.Runtime{}
		shoot := &gardenertypes.Shoot{}
		var session e2e.ScenarioSession
		var clusterInSession e2e.ClusterInSession

		By("Given an SKR instance exists", func() {
			nsList := &corev1.NamespaceList{}
			err := infra.Garden().Client().List(infra.Ctx(), nsList)
			Expect(err).NotTo(HaveOccurred())

			id, err := world.Keb().CreateInstance(infra.Ctx(),
				e2ekeb.WithAlias(alias),
				e2ekeb.WithProvider("gcp"),
			)
			Expect(err).NotTo(HaveOccurred())
			instanceDetails = id

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.Garden().Client(), shoot,
					NewObjActions(WithName(instanceDetails.ShootName), WithNamespace(config.GardenNamespace)),
					sim.HavingShootReady,
				).
				WithTimeout(8 * time.Second).
				Should(Succeed())

			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(), infra.KCP().Client(), rt,
					NewObjActions(WithName(instanceDetails.RuntimeID)),
					sim.HavingRuntimeProvisioningCompleted,
				).
				Should(Succeed())
		})

		By("And Given Scenario Session is created", func() {
			session = e2e.NewScenarioSession(world)
		})

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
			Expect(c.ClusterAlias()).To(Equal(alias))
		})

		By("When session is terminated", func() {
			err := session.Terminate(infra.Ctx())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then SKR instance still exists", func() {
			arr, err := world.Keb().List(infra.Ctx(), e2ekeb.WithAlias(alias))
			Expect(err).NotTo(HaveOccurred())
			Expect(arr).To(HaveLen(1))
			Expect(arr[0].Alias).To(Equal(alias))
		})

		By("// cleanup: delete SKR instance", func() {
			err := world.Keb().DeleteInstance(infra.Ctx(), e2ekeb.WithRuntime(rt.Name))
			Expect(err).NotTo(HaveOccurred())

			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), rt).
				Should(Succeed())
		})
	})

	It("Scenario: Add new runtime deletes SKR instances on session terminate", func() {

		alias := "d5c23067-6d8c-4538-a52a-f76dda4cdafb"

		rt := &infrastructuremanagerv1.Runtime{}
		shoot := &gardenertypes.Shoot{}

		session := e2e.NewScenarioSession(world)
		var clusterInSession e2e.ClusterInSession

		By("When Scenario session CreateNewSkrCluster() is called", func() {
			cis, err := session.CreateNewSkrCluster(infra.Ctx(),
				e2ekeb.WithAlias(alias),
				e2ekeb.WithGlobalAccount("038bad72-d2a3-4659-9d9c-49e2de840162"),
				e2ekeb.WithSubAccount("064ed161-6034-4905-a243-d994c392b683"),
				e2ekeb.WithProvider("gcp"),
			)
			Expect(err).NotTo(HaveOccurred())
			clusterInSession = cis
		})

		By("Then SKR instance exists", func() {
			id, err := world.Keb().GetInstance(infra.Ctx(), clusterInSession.RuntimeID())
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
			shoot.Name = clusterInSession.ShootName()
			shoot.Namespace = config.GardenNamespace
			Expect(LoadAndCheck(
				infra.Ctx(), infra.Garden().Client(), shoot,
				NewObjActions(),
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
			Expect(c.ClusterAlias()).To(Equal(alias))
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
			arr, err := world.Keb().List(infra.Ctx(), e2ekeb.WithAlias(alias))
			Expect(err).NotTo(HaveOccurred())
			Expect(arr).To(HaveLen(0))
		})

		By("And Then Shoot does not exist", func() {
			Expect(IsDeleted(infra.Ctx(), infra.Garden().Client(), shoot)).
				To(Succeed())
		})
	})
})
