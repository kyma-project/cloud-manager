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

func createScenario[T client.Object](clientFn func() client.Client, title string, b Builder[T], ok bool, errMsg string, focus bool) {
	handler := func() {
		obj := b.Build()
		obj.SetName(uuid.NewString())
		dsl.SetDefaultNamespace(obj)

		err := clientFn().Create(infra.Ctx(), obj)
		if ok {
			Expect(err).NotTo(HaveOccurred(), title)
			_ = clientFn().Delete(infra.Ctx(), obj)
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

func updateScenario[T client.Object](clientFn func() client.Client, title string, b Builder[T], cb func(b Builder[T]), ok bool, errMsg string, focus bool) {
	handler := func() {
		obj := b.Build()
		obj.SetName(uuid.NewString())
		dsl.SetDefaultNamespace(obj)
		err := clientFn().Create(infra.Ctx(), obj)
		Expect(err).NotTo(HaveOccurred())
		err = clientFn().Get(infra.Ctx(), client.ObjectKeyFromObject(obj), obj)
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			_ = clientFn().Delete(infra.Ctx(), obj)
		}()
		cb(b)
		err = clientFn().Update(infra.Ctx(), obj)
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
	createScenario(func() client.Client { return infra.KCP().Client() },
		title, b, true, "", false)
}
func canNotCreate[T client.Object](title string, b Builder[T], errMsg string) {
	createScenario(func() client.Client { return infra.KCP().Client() },
		title, b, false, errMsg, false)
}
func canNotChange[T client.Object](title string, b Builder[T], cb func(b Builder[T]), errMsg string) {
	updateScenario(func() client.Client { return infra.KCP().Client() },
		title, b, cb, false, errMsg, false)
}

func canCreateSkr[T client.Object](title string, b Builder[T]) {
	createScenario(func() client.Client { return infra.SKR().Client() }, title, b, true, "", false)
}
func canNotCreateSkr[T client.Object](title string, b Builder[T], errMsg string) {
	createScenario(func() client.Client { return infra.SKR().Client() }, title, b, false, errMsg, false)
}
func canChangeSkr[T client.Object](title string, b Builder[T], cb func(b Builder[T])) {
	updateScenario(func() client.Client { return infra.SKR().Client() }, title, b, cb, true, "", false)
}
func canNotChangeSkr[T client.Object](title string, b Builder[T], cb func(b Builder[T]), errMsg string) {
	updateScenario(func() client.Client { return infra.SKR().Client() }, title, b, cb, false, errMsg, false)
}
