package util

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func NewUiCmListUnstructured() *unstructured.UnstructuredList {
	u := &unstructured.UnstructuredList{}
	u.SetAPIVersion("v1")
	u.SetKind("ConfigMapList")

	return u
}

func NewUiCmUnstructured() *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("v1")
	u.SetKind("ConfigMap")
	u.SetNamespace("kyma-system")

	return u

}
