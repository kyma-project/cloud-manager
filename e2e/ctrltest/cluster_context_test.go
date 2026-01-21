package ctrltest

import (
	"fmt"

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

		By("Then expression `cmOne` returns nil", func() {
			evaluator, err := e2e.NewEvaluatorBuilder().
				Add(world.Kcp()).
				Build(infra.Ctx())
			Expect(err).NotTo(HaveOccurred())

			v, err := evaluator.Eval("cmOne")
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(BeNil())
		})

		By("Then expression `cmTwo` returns nil", func() {
			evaluator, err := e2e.NewEvaluatorBuilder().
				Add(world.Kcp()).
				Build(infra.Ctx())
			Expect(err).NotTo(HaveOccurred())

			v, err := evaluator.Eval("cmTwo")
			Expect(err).NotTo(HaveOccurred())
			Expect(v).To(BeNil())
		})

		By("When cmTwo is created", func() {
			err := world.Kcp().GetClient().Create(infra.Ctx(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmTwoName,
					Namespace: config.KcpNamespace,
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
					Namespace: config.KcpNamespace,
				},
				Data: map[string]string{
					"cmTwoName": cmTwoName,
				},
			})
			Expect(err).NotTo(HaveOccurred())
			world.Kcp().GetCache().WaitForCacheSync(infra.Ctx())
		})

		By("Then  expression cmOne.metadata.name returns value", func() {
			Eventually(func() error {
				evaluator, err := e2e.NewEvaluatorBuilder().
					Add(world.Kcp()).
					Build(infra.Ctx())
				if err != nil {
					return fmt.Errorf("error creating evaluator: %w", err)
				}

				v, err := evaluator.Eval("cmOne.metadata.name")
				if err != nil {
					return fmt.Errorf("error evaluating cmOne.metadata.name: %w", err)
				}
				if v != cmOneName {
					return fmt.Errorf("expected cmOne.metadata.name to evaluate to %q, got %v", cmOneName, v)
				}
				return nil
			}).Should(Succeed())
		})

		By("And Then expression cmOne.data.cmTwoName returns value", func() {
			Eventually(func() error {
				evaluator, err := e2e.NewEvaluatorBuilder().
					Add(world.Kcp()).
					Build(infra.Ctx())
				if err != nil {
					return fmt.Errorf("error creating evaluator: %w", err)
				}

				v, err := evaluator.Eval("cmOne.data.cmTwoName")
				if err != nil {
					return fmt.Errorf("error evaluating cmOne.data.cmTwoName: %w", err)
				}
				if v != cmTwoName {
					return fmt.Errorf("expected cmOne.data.cmTwoName to evaluate to %q, got %v", cmTwoName, v)
				}
				return nil
			}).Should(Succeed())
		})

		By("And Then expression cmTwo.metadata.name returns value", func() {
			Eventually(func() error {
				evaluator, err := e2e.NewEvaluatorBuilder().
					Add(world.Kcp()).
					Build(infra.Ctx())
				if err != nil {
					return fmt.Errorf("error creating evaluator: %w", err)
				}

				v, err := evaluator.Eval("cmTwo.metadata.name")
				if err != nil {
					return fmt.Errorf("error evaluating cmTwo.metadata.name: %w", err)
				}
				if v != cmTwoName {
					return fmt.Errorf("expected cmTwo.metadata.name to evaluate to %q, got %v", cmTwoName, v)
				}
				return nil
			}).Should(Succeed())
		})

		By("// cleanup: delete cmOne", func() {
			err := world.Kcp().GetClient().Delete(infra.Ctx(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmOneName,
					Namespace: config.KcpNamespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("// cleanup: delete cmTwo", func() {
			err := world.Kcp().GetClient().Delete(infra.Ctx(), &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmTwoName,
					Namespace: config.KcpNamespace,
				},
			})
			Expect(err).NotTo(HaveOccurred())
		})
	})

})
