package cloudcontrol

import (
	gardenertypes "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	"github.com/kyma-project/cloud-manager/api"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	"github.com/kyma-project/cloud-manager/pkg/external/infrastructuremanagerv1"
	kcpsubscription "github.com/kyma-project/cloud-manager/pkg/kcp/subscription"
	. "github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: Runtime", func() {

	It("Scenario: Runtime is created and deleted", func() {

		name := "75c7cc86-443e-444e-aa0b-b65e8296a8a0"
		shootName := "some-shoot-ac22f"
		secretBindingName := "my-secret-binding"

		var runtime *infrastructuremanagerv1.Runtime
		subscription := &cloudcontrolv1beta1.Subscription{}

		kcpsubscription.Ignore.AddName(secretBindingName)

		By("When Runtime is created", func() {
			runtime = &infrastructuremanagerv1.Runtime{
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: infra.KCP().Namespace(),
					Labels: map[string]string{
						cloudcontrolv1beta1.LabelRuntimeId:            name,
						cloudcontrolv1beta1.LabelScopeGlobalAccountId: "6329e93d-591f-4b1e-83ed-3dc6f9f426d7",
						cloudcontrolv1beta1.LabelScopeSubaccountId:    "f6d42db7-1195-4ff5-9787-0edb471c75cb",
						cloudcontrolv1beta1.LabelScopeShootName:       "some-shoot-ac22f",
						cloudcontrolv1beta1.LabelKymaName:             name,
						cloudcontrolv1beta1.LabelScopeBrokerPlanName:  "aws",
						cloudcontrolv1beta1.LabelScopeRegion:          "eu-west-1",
					},
				},
				Spec: infrastructuremanagerv1.RuntimeSpec{
					Security: infrastructuremanagerv1.Security{
						Administrators: []string{"someone@sap.com"},
					},
					Shoot: infrastructuremanagerv1.RuntimeShoot{
						Name: shootName,
						Provider: infrastructuremanagerv1.Provider{
							Type: "aws", // required!!!
							Workers: []gardenertypes.Worker{
								{
									Name: "worker1",
									Machine: gardenertypes.Machine{
										Image: &gardenertypes.ShootMachineImage{
											Name: "gardenlinux",
										},
										Type: "m5.large",
									},
								},
							},
						},
						SecretBindingName: secretBindingName,
					},
				},
			}

			Expect(CreateObj(infra.Ctx(), infra.KCP().Client(), runtime)).
				To(Succeed())

			_, err := composed.PatchObjAddFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, runtime, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then Subscription is created", func() {
			Eventually(LoadAndCheck).
				WithArguments(infra.Ctx(), infra.KCP().Client(), subscription, NewObjActions(WithName(secretBindingName))).
				Should(Succeed())
		})

		By("And Then Subscription has labels as Runtime", func() {
			for _, labelName := range cloudcontrolv1beta1.ScopeLabels {
				rVal, ok := runtime.Labels[labelName]
				Expect(ok).To(BeTrue(), "unexepcted logical error - runtime should have label %s set", labelName)
				sVal, ok := subscription.Labels[labelName]
				Expect(ok).To(BeTrue(), "subscription should have label %s", labelName)
				Expect(sVal).To(Equal(rVal), "subscription should have label %s with value %q, but it has %q", labelName, rVal, sVal)
			}
		})

		// DELETE ===============================================================

		By("When Runtime is deleted", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), runtime)).To(Succeed())
		})

		By("// cleanup: delete Subscription", func() {
			Expect(Delete(infra.Ctx(), infra.KCP().Client(), subscription)).
				To(Succeed())
			Expect(IsDeleted(infra.Ctx(), infra.KCP().Client(), subscription)).
				To(Succeed())
		})

		By("// cleanup: delete Runtime", func() {
			_, err := composed.PatchObjRemoveFinalizer(infra.Ctx(), api.CommonFinalizerDeletionHook, runtime, infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
			Expect(IsDeleted(infra.Ctx(), infra.KCP().Client(), runtime)).
				To(Succeed())
		})
	})

})
