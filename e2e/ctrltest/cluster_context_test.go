package ctrltest

import (
	"time"

	"github.com/kyma-project/cloud-manager/e2e"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: Cluster context", func() {

	It("Scenario: resource declaration with dependencies", func() {

		const (
			cmOneName = "25ca5214-2213-4597-b7ed-8e1892204a38"
			cmTwoName = "f2622fb5-8c28-4c48-9601-6d6595f71e47"
		)

		By("Given cmOne ConfigMap does not exist")

		By("And Given cmTwo ConfigMap does not exist")

		By("And Given cmOne resource is declared", func() {
			err := world.Kcp().AddResources(infra.Ctx(), &e2e.ResourceDeclaration{
				Alias:      "cmOne",
				Kind:       "ConfigMap",
				ApiVersion: "v1",
				Name:       cmOneName,
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("And Given cmTwo resource is declared", func() {
			err := world.Kcp().AddResources(infra.Ctx(), &e2e.ResourceDeclaration{
				Alias:      "cmTwo",
				Kind:       "ConfigMap",
				ApiVersion: "v1",
				Name:       "${cmOne.data.cmTwoName}",
			})
			Expect(err).NotTo(HaveOccurred())
		})

		var evaluator e2e.Evaluator

		By("When Evaluation context is created", func() {
			e, err := e2e.NewEvaluatorBuilder().
				Add(world.Kcp()).
				Build(infra.Ctx())
			Expect(err).NotTo(HaveOccurred())
			evaluator = e
		})

		By("Then expression `cmOne` returns nil", func() {
			v, err := evaluator.Eval("cmOne")
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(BeNil())
		})

		By("Then expression `cmTwo` returns nil", func() {
			v, err := evaluator.Eval("cmTwo")
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(BeNil())
		})

		By("When cmTwo is created", func() {
			err := world.Kcp().GetClient().Create(infra.Ctx(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmTwoName,
					Namespace: "default",
				},
				Data: map[string]string{
					"myName": "cmTwo",
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("And When cmOne is created", func() {
			err := world.Kcp().GetClient().Create(infra.Ctx(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmOneName,
					Namespace: "default",
				},
				Data: map[string]string{
					"cmTwoName": cmTwoName,
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("When Evaluation context is created", func() {
			// give time to cache to refresh
			time.Sleep(2 * time.Second)
			e, err := e2e.NewEvaluatorBuilder().
				Add(world.Kcp()).
				Build(infra.Ctx())
			Expect(err).NotTo(HaveOccurred())
			evaluator = e
		})

		By("Then expression cmOne.metadata.name returns value", func() {
			v, err := evaluator.Eval("cmOne.metadata.name")
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(Equal(cmOneName))
		})

		By("Then expression cmOne.data.cmTwoName returns value", func() {
			v, err := evaluator.Eval("cmOne.data.cmTwoName")
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(Equal(cmTwoName))
		})

		By("And Then expression cmTwo.metadata.name returns value", func() {
			v, err := evaluator.Eval("cmTwo.metadata.name")
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(Equal(cmTwoName))
		})

		By("// cleanup: delete cmOne", func() {
			err := world.Kcp().GetClient().Delete(infra.Ctx(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmOneName,
					Namespace: "default",
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("// cleanup: delete cmTwo", func() {
			err := world.Kcp().GetClient().Delete(infra.Ctx(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmTwoName,
					Namespace: "default",
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
