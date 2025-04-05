package objkind

import (
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"testing"
)

func NewUnstructuredWithGVK(g, v, k string) *unstructured.Unstructured {
	x := &unstructured.Unstructured{}
	x.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   g,
		Version: v,
		Kind:    k,
	})
	return x
}

func NewCrdUnstructuredWithKindGroup(t *testing.T, baseCrd *unstructured.Unstructured, k, g string) *unstructured.Unstructured {
	crd := baseCrd.DeepCopyObject().(*unstructured.Unstructured)
	err := unstructured.SetNestedField(crd.Object, k, "spec", "names", "kind")
	assert.NoError(t, err)
	err = unstructured.SetNestedField(crd.Object, g, "spec", "group")
	assert.NoError(t, err)
	return crd
}

func NewCrdTypedWithKindGroup(_ *testing.T, baseCrd *apiextensions.CustomResourceDefinition, k, g string) *apiextensions.CustomResourceDefinition {
	crd := baseCrd.DeepCopyObject().(*apiextensions.CustomResourceDefinition)
	crd.Spec.Names.Kind = k
	crd.Spec.Group = g
	return crd
}

func NewBusolaCmTypedKindGroup(t *testing.T, k string) *corev1.ConfigMap {
	data := map[string]interface{}{}
	err := unstructured.SetNestedField(data, "cloud-resources.kyma-project.io", "resource", "group")
	assert.NoError(t, err)
	err = unstructured.SetNestedField(data, k, "resource", "kind")
	assert.NoError(t, err)

	b, err := yaml.Marshal(data)
	assert.NoError(t, err)

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				"busola.io/extension": "resource",
			},
		},
		Data: map[string]string{
			"general": string(b),
		},
	}
}

func NewBusolaCmUnstructuredKindGroup(t *testing.T, k string) *unstructured.Unstructured {
	data := map[string]interface{}{}
	err := unstructured.SetNestedField(data, "cloud-resources.kyma-project.io", "resource", "group")
	assert.NoError(t, err)
	err = unstructured.SetNestedField(data, k, "resource", "kind")
	assert.NoError(t, err)

	b, err := yaml.Marshal(data)
	assert.NoError(t, err)

	u := &unstructured.Unstructured{Object: map[string]interface{}{}}
	u.SetAPIVersion("v1")
	u.SetKind("ConfigMap")
	u.SetLabels(map[string]string{
		"busola.io/extension": "resource",
	})
	err = unstructured.SetNestedMap(u.Object, map[string]interface{}{
		"general": string(b),
	}, "data")
	assert.NoError(t, err)

	return u
}
