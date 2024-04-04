package cr

import (
	"fmt"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("Feature: CloudResources module CR", func() {

	BeforeEach(func() {
		Expect(DeleteAllOfSKR(infra)).
			NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := infra.SKR().EnsureCrds(infra.Ctx())
		Expect(err).NotTo(HaveOccurred(), "failed ensuring CRDs are installed")
	})

	It("Scenario: CloudResources is created and deleted", func() {

		const (
			crName = "06da5d7d-72c2-4acf-b6dc-326043ba3eda"
		)
		cr := &cloudresourcesv1beta1.CloudResources{}

		By("When CloudResources is created", func() {
			Eventually(CreateNamespace).
				WithArguments(infra.Ctx(), infra.SKR().Client(), &corev1.Namespace{}, WithName("kyma-system")).
				Should(Succeed(), "failed creating kyma-system namespace")

			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr,
					WithNamespace("kyma-system"),
					WithName(crName),
				).
				Should(Succeed(), "failed creating CloudResources CR")
		})

		By("Then CloudResources has Ready condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr, NewObjActions(), HavingConditionTrue(cloudresourcesv1beta1.ConditionTypeReady)).
				Should(Succeed(), "expected CloudResources to have Ready condition, but it does not")
		})

		By("And Then CloudResources has status.state='Ready'", func() {
			Expect(cr.Status.State).To(Equal(cloudresourcesv1beta1.ModuleStateReady))
		})

		By("And Then CloudResources has finalizer", func() {
			Expect(controllerutil.ContainsFinalizer(cr, cloudresourcesv1beta1.Finalizer)).
				To(BeTrue(), "expected CloudResources to have finalizer, but it does not")
		})

		By("And Then CloudResources has status.served='true'", func() {
			Expect(cr.Status.Served).To(Equal(cloudresourcesv1beta1.ServedTrue), "expected .status.served to be 'true'")
		})

		// DELETE ========================

		By("When CloudResources is deleted", func() {
			Eventually(Delete).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr).
				Should(Succeed(), "failed deleting CloudResources")
		})

		By("Then CloudResources should not exist", func() {
			Eventually(IsDeleted).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr).
				Should(Succeed(), "expected CloudResources not to exist, but it still exists")
		})

		By("And Then CRDs are uninstalled", func() {
			Eventually(func() error {
				list := util.NewCrdListUnstructured()
				err := infra.SKR().Client().List(infra.Ctx(), list)
				Expect(err).NotTo(HaveOccurred())

				unexpectedList := []string{
					"awsnfsvolumes.cloud-resources.kyma-project.io",
					"gcpnfsvolumes.cloud-resources.kyma-project.io",
					"ipranges.cloud-resources.kyma-project.io",
				}
				for _, item := range list.Items {
					for _, unexpected := range unexpectedList {
						if item.GetName() == unexpected {
							return fmt.Errorf("expected CRD %s to be deleted, but it still exists", unexpected)
						}
					}
				}
				return nil
			}).Should(Succeed())

		})
	})

})
