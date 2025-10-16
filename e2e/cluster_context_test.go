package e2e

import (
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Feature: Cluster context", func() {

	It("Scenario: Basic resource declaration with ConfigMap", func() {

		const (
			cmOneName = "25ca5214-2213-4597-b7ed-8e1892204a38"
			//cmTwoName = "f2622fb5-8c28-4c48-9601-6d6595f71e47"
		)

		By("Given cmOne ConfigMap does not exist")

		By("When cmOne resource is declared", func() {
			err := world.Kcp().AddResources(infra.Ctx(), &ResourceDeclaration{
				Alias:      "cmOne",
				Kind:       "ConfigMap",
				ApiVersion: "v1",
				Name:       cmOneName,
				Namespace:  "default",
			})
			Expect(err).NotTo(HaveOccurred())
		})

		By("Then Cluster.Has(cmOne) returns true", func() {
			Expect(world.Kcp().Has("cmOne")).To(BeTrue())
		})

		By("Then Cluster.Get(cmOne) returns nil", func() {
			x, err := world.Kcp().Get(infra.Ctx(), "cmOne")
			Expect(err).NotTo(HaveOccurred())
			Expect(x).To(BeNil())
		})

		By("When cmOne is created", func() {
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cmOneName,
					Namespace: "default",
				},
				Data: map[string]string{
					"alias": "cmOne",
				},
			}
			Expect(world.Kcp().GetClient().Create(infra.Ctx(), cm)).To(Succeed())
		})

		By("Then Cluster.Get(cmOne) returns configmap", func() {
			x, err := world.Kcp().Get(infra.Ctx(), "cmOne")
			Expect(err).NotTo(HaveOccurred())
			Expect(x).NotTo(BeNil())
			cm, ok := x.(*corev1.ConfigMap)
			Expect(ok).To(BeTrue())
			Expect(cm.Data["alias"]).To(Equal("cmOne"))
		})

		By("And Then Cluster.EvaluationContext returns map with one entry for cmOne", func() {
			data, err := world.Kcp().EvaluationContext(infra.Ctx())
			Expect(err).NotTo(HaveOccurred())
			Expect(data).NotTo(BeNil())
			Expect(data).To(HaveKey("cmOne"))
			Expect(data["cmOne"]).NotTo(BeNil())
			util.AssertUnstructuredString(data, "ConfigMap", "cmOne", "kind")
			util.AssertUnstructuredString(data, cmOneName, "cmOne", "metadata", "name")
			util.AssertUnstructuredString(data, "cmOne", "cmOne", "data", "alias")
		})
	})

})
