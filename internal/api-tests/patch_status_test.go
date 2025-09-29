package api_tests

import (
	cloudresourcesv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-resources/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: Patch status", func() {

	It("Scenario: Patch status can remove status condition", func() {
		obj := &cloudresourcesv1beta1.GcpNfsVolumeRestore{}
		name := "46ae569e-5e28-4cf5-b888-1d973e6bf4cb"

		By("When GcpNfsVolumeRestore is created", func() {
			Expect(CreateGcpNfsVolumeRestore(
				infra.Ctx(), infra.SKR().Client(), obj,
				WithName(name),
				WithRestoreSourceBackup(name),
				WithRestoreDestinationVolume(name),
			)).To(Succeed())
		})

		By("Then GcpNfsVolumeRestore has no conditions", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.SKR().Client(), obj, NewObjActions())).
				To(Succeed())
			Expect(obj.Status.Conditions).To(HaveLen(0))
		})

		By("When GcpNfsVolumeRestore is patched with Ready condition", func() {
			meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
				Type:    "Ready",
				Status:  metav1.ConditionTrue,
				Reason:  "Ready",
				Message: "Ready",
			})
			Expect(composed.PatchObjStatus(infra.Ctx(), obj, infra.SKR().Client())).
				To(Succeed())
		})

		By("Then GcpNfsVolumeRestore has Ready condition", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.SKR().Client(), obj, NewObjActions())).
				To(Succeed())
			Expect(meta.FindStatusCondition(obj.Status.Conditions, "Ready")).
				NotTo(BeNil())
		})

		By("When GcpNfsVolumeRestore is patched with Ready removed and Ready2 added conditions", func() {
			meta.RemoveStatusCondition(&obj.Status.Conditions, "Ready")
			meta.SetStatusCondition(&obj.Status.Conditions, metav1.Condition{
				Type:    "Ready2",
				Status:  metav1.ConditionTrue,
				Reason:  "Ready2",
				Message: "Ready2",
			})
			Expect(composed.PatchObjStatus(infra.Ctx(), obj, infra.SKR().Client())).
				To(Succeed())
		})

		By("Then GcpNfsVolumeRestore has Ready2 condition", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.SKR().Client(), obj, NewObjActions())).
				To(Succeed())
			Expect(meta.FindStatusCondition(obj.Status.Conditions, "Ready2")).
				NotTo(BeNil())
		})

		By("And Then GcpNfsVolumeRestore does not have Ready condition", func() {
			Expect(LoadAndCheck(infra.Ctx(), infra.SKR().Client(), obj, NewObjActions())).
				To(Succeed())
			Expect(meta.FindStatusCondition(obj.Status.Conditions, "Ready")).
				To(BeNil())
		})
	})
})
