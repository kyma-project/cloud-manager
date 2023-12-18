package util

import (
	apimachineryapi "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewKymaUnstructured() *apimachineryapi.Unstructured {
	u := &apimachineryapi.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "operator.kyma-project.io",
		Version: "v1beta1",
		Kind:    "Kyma",
	})
	return u
}
