package cr

import (
	"fmt"

	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	scopeprovider "github.com/kyma-project/cloud-manager/pkg/skr/common/scope/provider"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		scope := &cloudcontrolv1beta1.Scope{}

		//By("Given KCP Scope exists", func() {
		//	Expect(CreateScopeAws(infra.Ctx(), infra, scope, "aws-cr-acc-123",
		//		WithName(crName))).To(Succeed())
		//	meta.SetStatusCondition(scope.Conditions(), metav1.Condition{
		//		Type:    cloudcontrolv1beta1.ConditionTypeReady,
		//		Status:  metav1.ConditionTrue,
		//		Reason:  cloudcontrolv1beta1.ConditionTypeReady,
		//		Message: cloudcontrolv1beta1.ConditionTypeReady,
		//	})
		//	Expect(UpdateStatus(infra.Ctx(), infra.KCP().Client(), scope)).To(Succeed())
		//})

		By("Given KCP Scope does not exist", func() {
			infra.ScopeProvider().Add(scopeprovider.MatchingObjName(crName, infra.KCP().Namespace(), crName))
		})

		By("When CloudResources is created", func() {
			Eventually(CreateObj).
				WithArguments(infra.Ctx(), infra.SKR().Client(), cr,
					WithName(crName),
				).
				Should(Succeed(), "failed creating CloudResources CR")
		})

		By("Then CloudResources has MissingScope condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					cr,
					NewObjActions(),
					HavingCondition(cloudresourcesv1beta1.ConditionTypeReady, metav1.ConditionFalse, cloudresourcesv1beta1.ConditionReasonMissingScope, "Scope for SKR not found"),
				).
				Should(Succeed())
		})

		By("When KCP Scope is created", func() {
			// any provider can do
			Expect(CreateScopeAws(infra.Ctx(), infra, scope, crName,
				WithName(crName))).To(Succeed())
		})

		By("Then CloudResources has WaitingForDependency condition", func() {
			Eventually(LoadAndCheck).
				WithArguments(
					infra.Ctx(),
					infra.SKR().Client(),
					cr,
					NewObjActions(),
					HavingCondition(cloudresourcesv1beta1.ConditionTypeReady, metav1.ConditionUnknown, cloudresourcesv1beta1.ReasonWaitingForDependency, "Waiting Scope to become ready"),
				).
				Should(Succeed())
		})

		By("When KCP Scope is Ready", func() {
			meta.SetStatusCondition(scope.Conditions(), metav1.Condition{
				Type:    cloudcontrolv1beta1.ConditionTypeReady,
				Status:  metav1.ConditionTrue,
				Reason:  cloudcontrolv1beta1.ConditionTypeReady,
				Message: cloudcontrolv1beta1.ConditionTypeReady,
			})
			Expect(UpdateStatus(infra.Ctx(), infra.KCP().Client(), scope)).To(Succeed())
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
			Expect(controllerutil.ContainsFinalizer(cr, api.CommonFinalizerDeletionHook)).
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
