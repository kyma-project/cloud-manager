package tests

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	"github.com/kyma-project/cloud-manager/api"
	"github.com/kyma-project/cloud-manager/pkg/migrateFinalizers"
	"github.com/kyma-project/cloud-manager/pkg/util"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type objectList struct {
	withOld1 []*unstructured.Unstructured
	withOld2 []*unstructured.Unstructured
	withNew  []*unstructured.Unstructured
	without  []*unstructured.Unstructured
}

func (l *objectList) apply(clnt client.Client) {
	for _, obj := range l.withOld1 {
		err := clnt.Create(infra.Ctx(), obj)
		Expect(err).ToNot(HaveOccurred())
	}
	for _, obj := range l.withOld2 {
		err := clnt.Create(infra.Ctx(), obj)
		Expect(err).ToNot(HaveOccurred())
	}
	for _, obj := range l.withNew {
		err := clnt.Create(infra.Ctx(), obj)
		Expect(err).ToNot(HaveOccurred())
	}
	for _, obj := range l.without {
		err := clnt.Create(infra.Ctx(), obj)
		Expect(err).ToNot(HaveOccurred())
	}
}

func (l *objectList) loadObj(original *unstructured.Unstructured, clnt client.Client) *unstructured.Unstructured {
	o := &unstructured.Unstructured{}
	o.SetAPIVersion(original.GetAPIVersion())
	o.SetKind(original.GetKind())
	o.SetNamespace(original.GetNamespace())
	o.SetName(original.GetName())
	err := clnt.Get(infra.Ctx(), client.ObjectKeyFromObject(o), o)
	Expect(err).ToNot(HaveOccurred())
	return o
}

func (l *objectList) assertHasFinalizer(o *unstructured.Unstructured) {
	Expect(controllerutil.ContainsFinalizer(o, api.DO_NOT_USE_OLD_KcpFinalizer)).To(BeFalse(), fmt.Sprintf("kind %s is not expected to have oldFinalizer1", o.GetKind()))
	Expect(controllerutil.ContainsFinalizer(o, api.DO_NOT_USE_OLD_SkrFinalizer)).To(BeFalse(), fmt.Sprintf("kind %s is not expected to have oldFinalizer2", o.GetKind()))
	Expect(controllerutil.ContainsFinalizer(o, api.CommonFinalizerDeletionHook)).To(BeTrue(), fmt.Sprintf("kind %s is expected to have CommonFinalizerDeletionHook", o.GetKind()))
}

func (l *objectList) assertDoesNotHaveFinalizer(o *unstructured.Unstructured) {
	Expect(controllerutil.ContainsFinalizer(o, api.DO_NOT_USE_OLD_KcpFinalizer)).To(BeFalse())
	Expect(controllerutil.ContainsFinalizer(o, api.DO_NOT_USE_OLD_SkrFinalizer)).To(BeFalse())
	Expect(controllerutil.ContainsFinalizer(o, api.CommonFinalizerDeletionHook)).To(BeFalse())
}

func (l *objectList) assertFinalizersAreMigrated(clnt client.Client) {
	for _, x := range l.withOld1 {
		o := l.loadObj(x, clnt)
		err := clnt.Get(infra.Ctx(), client.ObjectKeyFromObject(o), o)
		Expect(err).ToNot(HaveOccurred())
		l.assertHasFinalizer(o)
	}
	for _, x := range l.withOld2 {
		o := l.loadObj(x, clnt)
		err := clnt.Get(infra.Ctx(), client.ObjectKeyFromObject(o), o)
		Expect(err).ToNot(HaveOccurred())
		l.assertHasFinalizer(o)
	}
	for _, x := range l.withNew {
		o := l.loadObj(x, clnt)
		err := clnt.Get(infra.Ctx(), client.ObjectKeyFromObject(o), o)
		Expect(err).ToNot(HaveOccurred())
		l.assertHasFinalizer(o)
	}
	for _, x := range l.without {
		o := l.loadObj(x, clnt)
		err := clnt.Get(infra.Ctx(), client.ObjectKeyFromObject(o), o)
		Expect(err).ToNot(HaveOccurred())
		l.assertDoesNotHaveFinalizer(o)
	}
}

var _ = Describe("Feature: Finalizer migration", func() {

	loadFixture := func(prefix string) []*unstructured.Unstructured {
		b, err := os.ReadFile(fmt.Sprintf("./%s_fixture_test.yaml", prefix))
		Expect(err).ToNot(HaveOccurred())
		list, _ := util.YamlMultiDecodeToUnstructured(b)
		return list
	}

	getObjects := func(prefix string, namespace string) *objectList {
		list := loadFixture(prefix)
		result := &objectList{
			withOld1: make([]*unstructured.Unstructured, 0, len(list)),
			withOld2: make([]*unstructured.Unstructured, 0, len(list)),
			withNew:  make([]*unstructured.Unstructured, 0, len(list)),
			without:  make([]*unstructured.Unstructured, 0, len(list)),
		}
		for _, u := range list {
			o1 := u.DeepCopy()
			o1.SetNamespace(namespace)
			o1.SetName(uuid.NewString())
			controllerutil.AddFinalizer(o1, api.DO_NOT_USE_OLD_KcpFinalizer)
			result.withOld1 = append(result.withOld1, o1)

			o2 := u.DeepCopy()
			o2.SetNamespace(namespace)
			o2.SetName(uuid.NewString())
			controllerutil.AddFinalizer(o2, api.DO_NOT_USE_OLD_SkrFinalizer)
			result.withOld2 = append(result.withOld2, o2)

			o3 := u.DeepCopy()
			o3.SetNamespace(namespace)
			o3.SetName(uuid.NewString())
			controllerutil.AddFinalizer(o3, api.CommonFinalizerDeletionHook)
			result.withNew = append(result.withNew, o3)

			o4 := u.DeepCopy()
			o4.SetNamespace(namespace)
			o4.SetName(uuid.NewString())
			result.without = append(result.without, o4)
		}

		return result
	}

	makeKyma := func() *unstructured.Unstructured {
		u := util.NewKymaUnstructured()
		u.SetName(uuid.NewString())
		u.SetNamespace("kcp-system")

		err := infra.KCP().Client().Create(infra.Ctx(), u)
		Expect(err).ToNot(HaveOccurred())
		return u
	}

	It("Scenario: SKR finalizer migration", func() {
		list := getObjects("skr", "default")
		list.apply(infra.SKR().Client())

		kyma := makeKyma()

		// Infra client is made w/out cache, so we can pass that single instance both as reader and writter
		// ControllerRuntime on the other hand, has specially crafter Client with cache options that can not
		// read until all controllers are started, and thus it also provides a separate client instance
		// configured w/out cache but exposes it only under the Reader interface
		mig := migrateFinalizers.NewMigrationForSkr(kyma.GetName(), infra.KCP().Client(), infra.KCP().Client(), infra.SKR().Client(), infra.SKR().Client(), logr.Discard())
		alreadyExecuted, err := mig.Run(infra.Ctx())
		Expect(err).ToNot(HaveOccurred())
		Expect(alreadyExecuted).To(BeFalse())

		alreadyExecuted, err = mig.Run(infra.Ctx())
		Expect(err).ToNot(HaveOccurred())
		Expect(alreadyExecuted).To(BeTrue())

		list.assertFinalizersAreMigrated(infra.SKR().Client())

		// assert kyma is annotated due to success
		kyma = list.loadObj(kyma, infra.KCP().Client())
		value, hasAnnotation := kyma.GetAnnotations()[migrateFinalizers.SuccessAnnotation]
		Expect(hasAnnotation).To(BeTrue())
		Expect(value).To(Equal("true"))
	})

	It("Scenario: KCP finalizer migration", func() {
		list := getObjects("kcp", "kcp-system")
		list.apply(infra.KCP().Client())

		// Infra client is made w/out cache, so we can pass that single instance both as reader and writter
		// ControllerRuntime on the other hand, has specially crafter Client with cache options that can not
		// read until all controllers are started, and thus it also provides a separate client instance
		// configured w/out cache but exposes it only under the Reader interface
		mig := migrateFinalizers.NewMigrationForKcp(infra.KCP().Client(), infra.KCP().Client(), logr.Discard())
		alreadyExecuted, err := mig.Run(infra.Ctx())
		Expect(err).ToNot(HaveOccurred())
		Expect(alreadyExecuted).To(BeFalse())

		alreadyExecuted, err = mig.Run(infra.Ctx())
		Expect(err).ToNot(HaveOccurred())
		Expect(alreadyExecuted).To(BeTrue())

		list.assertFinalizersAreMigrated(infra.KCP().Client())

		// assert kcp ConfigMap is created due to success
		cm := &corev1.ConfigMap{}
		err = infra.KCP().Client().Get(infra.Ctx(), client.ObjectKey{Namespace: "kcp-system", Name: migrateFinalizers.KcpConfigMapName}, cm)
		Expect(err).ToNot(HaveOccurred())
	})
})
