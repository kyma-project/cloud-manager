package sim

import (
	"fmt"

	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: KCP SIM", func() {

	It("Scenario: Instance is created, module added and removed, and instance deleted", func() {
		const (
			alias         = "test"
			globalAccount = "6d0e57f0-66ed-4e60-9eb6-346e3354c5d0"
			subAccount    = "4d08684f-86f1-4e55-885f-0ae6920e65fb"
			provider      = cloudcontrolv1beta1.ProviderAws
		)

		var instanceDetails e2ekeb.InstanceDetails
		rt := &infrastructuremanagerv1.Runtime{}
		shoot := &gardenertypes.Shoot{}

		kubeconfigCallCount := skrKubeconfigProviderInstance.GetCallCount(shoot.Name)

		By("When instance is created", func() {
			id, err := kebInstance.CreateInstance(infra.Ctx(),
				e2ekeb.WithAlias(alias),
				e2ekeb.WithGlobalAccount(globalAccount),
				e2ekeb.WithSubAccount(subAccount),
				e2ekeb.WithProvider(provider),
			)
			Expect(err).NotTo(HaveOccurred())
			instanceDetails = id
		})

		By("Then Runtime is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), rt, NewObjActions(WithName(instanceDetails.RuntimeID), WithNamespace(config.KcpNamespace))).
				Should(Succeed())
		})

		By("And Then Shoot is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.Garden().Client(), shoot, NewObjActions(WithName(instanceDetails.ShootName), WithNamespace(config.GardenNamespace))).
				Should(Succeed())
		})

		By("And Then eventually Shoot is ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.Garden().Client(), shoot, NewObjActions(), HavingShootReady).
				Should(Succeed())
		})

		By("Then Runtime has ProvisioningCompleted true", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), rt, NewObjActions(), HavingRuntimeProvisioningCompleted).
				Should(Succeed())
		})

		gc := &infrastructuremanagerv1.GardenerCluster{}

		By("And Then GardenerCluster is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gc, NewObjActions(WithName(rt.Name), WithNamespace(rt.Namespace))).
				Should(Succeed())
		})

		By("And Then GardenerCluster has spec.shoot.name set", func() {
			Expect(gc.Spec.Shoot.Name).To(Equal(rt.Spec.Shoot.Name))
		})

		By("And Then GardenerCluster status state is Ready", func() {
			Expect(gc.Status.State).To(Equal(infrastructuremanagerv1.ReadyState))
		})

		By("And Then Garden kubeconfig is generated", func() {
			Expect(skrKubeconfigProviderInstance.GetCallCount(rt.Spec.Shoot.Name) > kubeconfigCallCount).To(BeTrue())
		})

		By("And Then SKR kubeconfig secret is created", func() {
			secret := &corev1.Secret{}
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), secret, NewObjActions(WithName(gc.Spec.Kubeconfig.Secret.Name), WithNamespace(gc.Spec.Kubeconfig.Secret.Namespace)))).
				To(Succeed())
			Expect(secret.Data).To(HaveKey(gc.Spec.Kubeconfig.Secret.Key))
		})
		kcpKyma := &operatorv1beta2.Kyma{}

		By("And Then KCP Kyma is crated", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpKyma, NewObjActions(WithName(rt.Name), WithNamespace(rt.Namespace))).
				Should(Succeed())
		})

		By("And Then KCP Kyma status state is Ready", func() {
			Expect(kcpKyma.Status.State).To(Equal(operatorshared.StateReady))
		})

		skrKyma := &operatorv1beta2.Kyma{}

		By("And Then SKR Kyma is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("And Then SKR Kyma status.state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions(), func(obj client.Object) error {
					x, ok := obj.(*operatorv1beta2.Kyma)
					if !ok {
						return fmt.Errorf("expected Kyma object but got %T", obj)
					}
					if x.Status.State != operatorshared.StateReady {
						return fmt.Errorf("expected Kyma with ready state, but it is %q", x.Status.State)
					}
					return nil
				}).
				Should(Succeed())

			Expect(skrKyma.Status.State).To(Equal(operatorshared.StateReady))
		})

		cr := &cloudresourcesv1beta1.CloudResources{}

		By("And Then CloudResources CR is created", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.SKR().Client(), cr, NewObjActions(WithName("default"), WithNamespace("kyma-system")))).
				To(Succeed())
		})

		By("When cloud-manager module is added to SKR Kyma spec", func() {
			skrKyma.Spec.Modules = []operatorv1beta2.Module{
				{
					Name: "cloud-manager",
				},
			}
			Expect(infra.SKR().Client().Update(infra.Ctx(), skrKyma)).
				To(Succeed())
		})

		By("Then cloud-manager module is present in KCP Kyma status", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpKyma, NewObjActions(), func(obj client.Object) error {
					x, ok := obj.(*operatorv1beta2.Kyma)
					if !ok {
						return fmt.Errorf("expected Kyma object but got %T", obj)
					}
					var cmModule operatorv1beta2.ModuleStatus
					found := false
					for _, m := range x.Status.Modules {
						if m.Name == "cloud-manager" {
							cmModule = m
							found = true
							break
						}
					}
					if !found {
						return fmt.Errorf("cloud-manager module not found in KCP Kyma status")
					}
					if cmModule.State != operatorshared.StateReady {
						return fmt.Errorf("expected KCP Kyma with cloud-manager module in ready state, but it is %q", cmModule.State)
					}
					return nil
				}).
				Should(Succeed())
		})

		By("When cloud-manager module is removed from SKR Kyma spec", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions())).
				To(Succeed())
			skrKyma.Spec.Modules = []operatorv1beta2.Module{}
			Expect(infra.SKR().Client().Update(infra.Ctx(), skrKyma)).
				To(Succeed())
		})

		By("Then cloud-manager module is not present in KCP Kyma status", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpKyma, NewObjActions(), func(obj client.Object) error {
					x, ok := obj.(*operatorv1beta2.Kyma)
					if !ok {
						return fmt.Errorf("expected Kyma object but got %T", obj)
					}
					found := false
					for _, m := range x.Status.Modules {
						if m.Name == "cloud-manager" {
							found = true
							break
						}
					}
					if found {
						return fmt.Errorf("cloud-manager module found in KCP Kyma status")
					}
					return nil
				}).
				Should(Succeed())
		})

		By("When Runtime is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), rt)).
				To(Succeed())
		})

		By("Then CloudResources is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr).
				Should(Succeed())
		})

		By("And Then SKR Kyma is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrKyma).
				Should(Succeed())
		})

		By("And Then KCP Kyma is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpKyma).
				Should(Succeed())
		})

		By("And Then Shoot is deleted", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.Garden().Client(), shoot).
				Should(Succeed())
		})

		By("And Then Runtime does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), rt).
				Should(Succeed())
		})

	})
})
