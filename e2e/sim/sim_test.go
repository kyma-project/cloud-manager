package sim

import (
	"fmt"
	"time"

	gardenerapicore "github.com/gardener/gardener/pkg/apis/core"
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	e2eclean "github.com/kyma-project/cloud-manager/e2e/clean"
	e2ekeb "github.com/kyma-project/cloud-manager/e2e/keb"
	commonscheme "github.com/kyma-project/cloud-manager/pkg/common/scheme"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorshared"
	"github.com/kyma-project/cloud-manager/pkg/external/operatorv1beta2"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: KCP SIM", Serial, func() {

	BeforeEach(func() {
		By("Given clusters are clean", func() {
			commonOps := []e2eclean.Option{
				e2eclean.WithWait(true),
				// intentionally low timeout, not waiting for something to remove finalizers, but doing a force delete
				e2eclean.WithTimeout(time.Millisecond),
				e2eclean.WithForceDeleteOnTimeout(true),
				e2eclean.WithLogger(composed.LoggerFromCtx(infra.Ctx())),
			}
			makeOpts := func(opts ...e2eclean.Option) []e2eclean.Option {
				return append(commonOps, opts...)
			}
			var err error

			err = e2eclean.Clean(infra.Ctx(), makeOpts(
				e2eclean.WithClient(infra.KCP().Client()),
				e2eclean.WithScheme(commonscheme.KcpScheme),
				e2eclean.WithMatchers(
					e2eclean.MatchingGroup(infrastructuremanagerv1.GroupVersion.Group),
					e2eclean.MatchingGroup(operatorv1beta2.GroupVersion.Group),
					e2eclean.MatchingGroup(cloudcontrolv1beta1.GroupVersion.Group),
				),
			)...)
			Expect(err).NotTo(HaveOccurred())

			err = e2eclean.Clean(infra.Ctx(), makeOpts(
				e2eclean.WithClient(infra.Garden().Client()),
				e2eclean.WithScheme(commonscheme.GardenScheme),
				e2eclean.WithMatchers(
					e2eclean.MatchAll(
						e2eclean.MatchingGroup(gardenerapicore.GroupName),
						e2eclean.MatchingKind("Shoot"),
					),
				),
			)...)
			Expect(err).NotTo(HaveOccurred())

			err = e2eclean.Clean(infra.Ctx(), makeOpts(
				e2eclean.WithClient(infra.SKR().Client()),
				e2eclean.WithScheme(commonscheme.SkrScheme),
				e2eclean.WithMatchers(
					e2eclean.MatchingGroup(operatorv1beta2.GroupVersion.Group),
					e2eclean.MatchingGroup(cloudresourcesv1beta1.GroupVersion.Group),
				),
			)...)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	It("Scenario: Instance is created, module added and removed, and instance deleted", Serial, func() {
		const (
			alias         = "test1"
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

		By("And Then Shoot is ready", func() {
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

		By("And Then GardenerCluster exists", func() {
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

		By("And Then SKR kubeconfig secret exists", func() {
			secret := &corev1.Secret{}
			Expect(LoadAndCheck(infra.Ctx(), infra.KCP().Client(), secret, NewObjActions(WithName(gc.Spec.Kubeconfig.Secret.Name), WithNamespace(gc.Spec.Kubeconfig.Secret.Namespace)))).
				To(Succeed())
			Expect(secret.Data).To(HaveKey(gc.Spec.Kubeconfig.Secret.Key))
		})
		kcpKyma := &operatorv1beta2.Kyma{}

		By("And Then KCP Kyma exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpKyma, NewObjActions(WithName(rt.Name), WithNamespace(rt.Namespace))).
				Should(Succeed())
		})

		By("And Then KCP Kyma status state is Ready", func() {
			Expect(kcpKyma.Status.State).To(Equal(operatorshared.StateReady))
		})

		skrKyma := &operatorv1beta2.Kyma{}

		By("And Then SKR Kyma exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("And Then SKR Kyma status.state is Ready", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions(), HavingState(string(operatorshared.StateReady))).
				Should(Succeed())

			Expect(skrKyma.Status.State).To(Equal(operatorshared.StateReady))
		})

		By("When module cloud-manager is enabled", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions())).
				To(Succeed())
			skrKyma.Spec.Modules = []operatorv1beta2.Module{{Name: "cloud-manager"}}
			Expect(infra.SKR().Client().Update(infra.Ctx(), skrKyma)).
				To(Succeed())
		})

		cr := &cloudresourcesv1beta1.CloudResources{}

		By("Then CloudResources CR exists", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr, NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("# CloudManager when enabled adds finalizer to CloudResources CR")

		By("And Then finalizer is added to CloudResources CR ", func() {
			_, err := composed.PatchObjAddFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, cr, infra.SKR().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("# CloudManager when enabled sets CloudResources CR state to Ready")

		By("And Then CloudResources CR have Ready state", func() {
			cr.Status.State = "Ready"
			err := composed.PatchObjStatus(infra.Ctx(), cr, infra.SKR().Client())
			Expect(err).NotTo(HaveOccurred())
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

		// delete ============

		By("When cloud-manager module is removed from SKR Kyma spec", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions())).
				To(Succeed())
			skrKyma.Spec.Modules = []operatorv1beta2.Module{}
			Expect(infra.SKR().Client().Update(infra.Ctx(), skrKyma)).
				To(Succeed())
		})

		By("Then CloudResources CR has deletion timestamp", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed())
		})

		By("And Then CloudResources CR finalizer is removed", func() {
			_, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, cr, infra.SKR().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("And Then cloud-manager module is not present in KCP Kyma status", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpKyma, NewObjActions(), NotHavingKymaModuleInStatus("cloud-manager")).
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

	It("Scenario: Instance is deleted with existing resources", Serial, func() {
		const (
			alias    = "test2"
			provider = cloudcontrolv1beta1.ProviderAws
		)

		var instanceDetails e2ekeb.InstanceDetails
		rt := &infrastructuremanagerv1.Runtime{}
		shoot := &gardenertypes.Shoot{}

		By("Given SKR instance is created", func() {
			id, err := kebInstance.CreateInstance(infra.Ctx(),
				e2ekeb.WithAlias(alias),
				e2ekeb.WithProvider(provider),
			)
			Expect(err).NotTo(HaveOccurred())
			instanceDetails = id
		})

		By("# SIM creates Runtime, Shoot, GardenerCluster, KCP Kyma, SKR Kyma")

		By("And Given Runtime is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), rt, NewObjActions(WithName(instanceDetails.RuntimeID), WithNamespace(config.KcpNamespace))).
				Should(Succeed())
		})

		By("And Given Shoot is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.Garden().Client(), shoot, NewObjActions(WithName(instanceDetails.ShootName), WithNamespace(config.GardenNamespace))).
				Should(Succeed())
		})

		gc := &infrastructuremanagerv1.GardenerCluster{}

		By("And Given GardenerCluster is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), gc, NewObjActions(WithName(rt.Name), WithNamespace(rt.Namespace))).
				Should(Succeed())
		})

		kcpKyma := &operatorv1beta2.Kyma{}

		By("And Given KCP Kyma is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpKyma, NewObjActions(WithName(rt.Name), WithNamespace(rt.Namespace))).
				Should(Succeed())
		})

		skrKyma := &operatorv1beta2.Kyma{}

		By("And Given SKR Kyma is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("And Given SKR Kyma status.state is Ready", func() {
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

		By("# User activates cloud-manager module by adding to SKR Kyma spec")

		By("And Given module cloud-manager is enabled", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions())).
				To(Succeed())
			skrKyma.Spec.Modules = []operatorv1beta2.Module{
				{
					Name: "cloud-manager",
				},
			}
			Expect(infra.SKR().Client().Update(infra.Ctx(), skrKyma)).
				To(Succeed())
		})

		cr := &cloudresourcesv1beta1.CloudResources{}

		By("# SIM creates CloudResources CR")

		By("And Given CloudResources CR is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr, NewObjActions(WithName("default"), WithNamespace("kyma-system"))).
				Should(Succeed())
		})

		By("# CloudManager when enabled sets the finalizer on CloudResources CR")

		By("And Given CloudResources CR have finalizer", func() {
			_, err := composed.PatchObjAddFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, cr, infra.SKR().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("# CloudManager when enabled sets CloudResources CR state to Ready")

		By("And Given CloudResources CR have Ready state", func() {
			cr.Status.State = "Ready"
			err := composed.PatchObjStatus(infra.Ctx(), cr, infra.SKR().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("# SIM updates CloudResources CR when observed its ready status state")

		By("Then CloudManager module has Ready status in SKR Kyma", func() {
			Eventually(func() error {
				if err := LoadAndCheck(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions()); err != nil {
					return err
				}
				m, ok := skrKyma.GetModuleStatusMap()["cloud-manager"]
				if !ok {
					return fmt.Errorf("cloud-manager module not present in SKR Kyma status")
				}
				if m.State != operatorshared.StateReady {
					return fmt.Errorf("expected ready state, but it is %q", m.State)
				}
				return nil
			}).Should(Succeed())
		})

		ipRange := &cloudresourcesv1beta1.IpRange{}

		By("When default IpRange is created", func() {
			ipRange.Name = "default"
			ipRange.Namespace = "kyma-system"
			ipRange.Finalizers = []string{api.CommonFinalizerDeletionHook}
			err := infra.SKR().Client().Create(infra.Ctx(), ipRange)
			Expect(err).NotTo(HaveOccurred())
		})

		By("And When SKR Instance is deleted", func() {
			err := kebInstance.DeleteInstance(infra.Ctx(), e2ekeb.WithRuntime(instanceDetails.RuntimeID), e2ekeb.WithTimeout(0))
			Expect(err).NotTo(HaveOccurred())
		})

		// TODO: fix kymaKcp reconciler so this commented steps work
		//By("# SIM has deleted the KCP Kyma since instance is being deleted")
		//
		//By("Then KCP Kyma has deletion timestamp", func() {
		//	Eventually(LoadAndCheck).
		//		WithArguments(infra.Ctx(), infra.SKR().Client(), kcpKyma, NewObjActions(), HavingDeletionTimestamp()).
		//		Should(Succeed())
		//})

		By("# SIM has deleted the SKR Kyma since KCP Kyma is being deleted")

		By("Then SKR Kyma has deletion timestamp", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed())
		})

		By("# SIM has deleted the CloudResources CR since SKR Kyma is being deleted")

		By("And Then CloudResources CR has deletion timestamp", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr, NewObjActions(), HavingDeletionTimestamp()).
				Should(Succeed())
		})

		By("# This timeout is the main subject being testing in this scenario!!!")

		By("When CloudResources CR is not deleted within the timeout period", func() {
			fakeClock.Step(timeoutRemoveModuleToErrorState + time.Minute)
		})

		By("Then SKR Kyma cloud-manager module has error status", func() {
			Eventually(func() error {
				if err := LoadAndCheck(infra.Ctx(), infra.SKR().Client(), skrKyma, NewObjActions()); err != nil {
					return fmt.Errorf("error loading SKR Kyma: %w", err)
				}

				m, ok := skrKyma.GetModuleStatusMap()["cloud-manager"]
				if !ok {
					return fmt.Errorf("cloud-manager module not present in SKR Kyma status")
				}
				if m.State != operatorshared.StateError {
					return fmt.Errorf("expected error state, but it is %q", m.State)
				}
				return nil
			})
		})

		By("When default IpRange is deleted", func() {
			err := infra.SKR().Client().Delete(infra.Ctx(), ipRange)
			Expect(err).NotTo(HaveOccurred())
			_, err = composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, ipRange, infra.SKR().Client())
			Expect(err).NotTo(HaveOccurred())

			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), ipRange).
				Should(Succeed())
		})

		By("# CloudManager removes finalizer from CloudResources CR since there are no more CM resources")

		By("And When CloudResources CR finalizer is removed ", func() {
			_, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, cr, infra.SKR().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("And When CloudResources CR does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr).
				Should(Succeed())
		})

		By("Then SKR Kyma does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), skrKyma).
				Should(Succeed())
		})

		By("And Then KCP Kyma does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.KCP().Client(), kcpKyma).
				Should(Succeed())
		})

		By("And Then Shoot does not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.Garden().Client(), shoot).
				Should(Succeed())
		})

		// TODO: fix this works
		By("And Then instance does not exist", func() {
			Eventually(func() error {
				id, err := kebInstance.GetInstance(infra.Ctx(), instanceDetails.RuntimeID)
				if err != nil {
					return fmt.Errorf("error getting instance details: %w", err)
				}
				if id != nil {
					return fmt.Errorf("instance still exists")
				}
				return nil
			}).Should(Succeed())
		})
	})

})
