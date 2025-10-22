package util

import (
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func AssertUnstructuredString(data map[string]interface{}, expected string, path ...string) {
	val, found, err := unstructured.NestedString(data, path...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(val).To(gomega.Equal(expected))
}

func AssertUnstructuredNil(data map[string]interface{}, path ...string) {
	val, found, err := unstructured.NestedFieldNoCopy(data, path...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(val).To(gomega.BeNil())
}
