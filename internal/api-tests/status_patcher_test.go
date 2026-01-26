package api_tests

import (
	"github.com/google/uuid"
	cloudcontrolv1beta1 "github.com/kyma-project/cloud-manager/api/cloud-control/v1beta1"
	"github.com/kyma-project/cloud-manager/pkg/composed"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ = Describe("Feature: Status Patcher", func() {

	createObj := func() (*cloudcontrolv1beta1.Subscription, error) {
		obj := &cloudcontrolv1beta1.Subscription{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: infra.KCP().Namespace(),
				Name:      uuid.NewString(),
			},
			Spec: cloudcontrolv1beta1.SubscriptionSpec{
				Details: cloudcontrolv1beta1.SubscriptionDetails{
					Garden: &cloudcontrolv1beta1.SubscriptionGarden{
						BindingName: "test",
					},
				},
			},
		}

		err := infra.KCP().Client().Create(infra.Ctx(), obj)
		if err != nil {
			return nil, err
		}

		err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(obj), obj)

		return obj, err
	}

	It("Scenario: Patch without changes", func() {

		var obj *cloudcontrolv1beta1.Subscription

		By("When object is created", func() {
			o, err := createObj()
			Expect(err).NotTo(HaveOccurred())
			obj = o
		})

		var version string
		var generation int64

		By("Then object has generation 1", func() {
			Expect(obj.GetGeneration()).NotTo(Equal(1))
			version = obj.ResourceVersion
			generation = obj.Generation
		})

		By("When object status is mutated and patched", func() {
			err := composed.NewStatusPatcher(obj).
				MutateStatus(func(x *cloudcontrolv1beta1.Subscription) {
					x.Status.Provider = cloudcontrolv1beta1.ProviderGCP
				}).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then object version has changed", func() {
			Expect(obj.ResourceVersion).NotTo(Equal(version))
			version = obj.ResourceVersion
		})

		By("Then object generation has NOT changed", func() {
			Expect(obj.Generation).To(Equal(generation))
		})

		By("When object status is patched without changes", func() {
			err := composed.NewStatusPatcher(obj).
				Patch(infra.Ctx(), infra.KCP().Client())
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then object version has NOT changed", func() {
			Expect(obj.ResourceVersion).To(Equal(version))
		})

		By("Then object generation also has NOT changed", func() {
			Expect(obj.Generation).To(Equal(generation))
		})

		By("// cleanup: delete object", func() {
			err := infra.KCP().Client().Delete(infra.Ctx(), obj)
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
