package util

import (
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func AssertUnstructuredString(data map[string]interface{}, expected string, path ...string) {
	val, found, err := unstructured.NestedMap(data, path...)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(found).To(gomega.BeTrue())
	gomega.Expect(val).To(gomega.Equal(expected))
}
