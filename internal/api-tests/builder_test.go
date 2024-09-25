package api_tests

import (
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/pkg/testinfra/dsl"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Builder[T client.Object] interface {
	Build() T
}

func createScenario[T client.Object](title string, b Builder[T], ok bool, errMsg string, focus bool) {
	handler := func() {
		obj := b.Build()
		obj.SetName(uuid.NewString())
		dsl.SetDefaultNamespace(obj)

		err := infra.KCP().Client().Create(infra.Ctx(), obj)
		if ok {
			Expect(err).NotTo(HaveOccurred(), title)
			_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
		} else {
			Expect(err).To(HaveOccurred(), title)
			if errMsg != "" {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		}
	}
	if focus {
		It("Scenario: "+title, Focus, handler)
	} else {
		It("Scenario: "+title, handler)
	}
}

func updateScenario[T client.Object](title string, b Builder[T], cb func(b Builder[T]), ok bool, errMsg string, focus bool) {
	handler := func() {
		obj := b.Build()
		obj.SetName(uuid.NewString())
		dsl.SetDefaultNamespace(obj)
		err := infra.KCP().Client().Create(infra.Ctx(), obj)
		Expect(err).NotTo(HaveOccurred())
		err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKeyFromObject(obj), obj)
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			_ = infra.KCP().Client().Delete(infra.Ctx(), obj)
		}()
		cb(b)
		err = infra.KCP().Client().Update(infra.Ctx(), obj)
		if ok {
			Expect(err).NotTo(HaveOccurred(), title)
		} else {
			Expect(err).To(HaveOccurred(), title)
			if errMsg != "" {
				Expect(err.Error()).To(ContainSubstring(errMsg))
			}
		}
	}
	if focus {
		It("Scenario: "+title, Focus, handler)
	} else {
		It("Scenario: "+title, handler)
	}
}

func canCreate[T client.Object](title string, b Builder[T]) {
	createScenario(title, b, true, "", false)
}
func canNotCreate[T client.Object](title string, b Builder[T], errMsg string) {
	createScenario(title, b, false, errMsg, false)
}
func canNotChange[T client.Object](title string, b Builder[T], cb func(b Builder[T]), errMsg string) {
	updateScenario(title, b, cb, false, errMsg, false)
}
