package util

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewGardenerClusterUnstructured() *unstructured.Unstructured {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructuremanager.kyma-project.io",
		Version: "v1",
		Kind:    "GardenerCluster",
	})
	return u
}

func NewGardenerClusterListUnstructured() *unstructured.UnstructuredList {
	u := &unstructured.UnstructuredList{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "infrastructuremanager.kyma-project.io",
		Version: "v1",
		Kind:    "GardenerClusterList",
	})
	return u
}

func ExtractGardenerClusterSummary(gc *unstructured.Unstructured) (*GardenerClusterSummary, error) {
	result := &GardenerClusterSummary{}
	if val, found, err := unstructured.NestedString(gc.Object, "spec", "kubeconfig", "secret", "key"); !found || err != nil {
		return nil, err
	} else {
		result.Key = val
	}
	if val, found, err := unstructured.NestedString(gc.Object, "spec", "kubeconfig", "secret", "name"); !found || err != nil {
		return nil, err
	} else {
		result.Name = val
	}
	if val, found, err := unstructured.NestedString(gc.Object, "spec", "kubeconfig", "secret", "namespace"); !found || err != nil {
		return nil, err
	} else {
		result.Namespace = val
	}
	if val, found, err := unstructured.NestedString(gc.Object, "spec", "shoot", "name"); !found || err != nil {
		return nil, err
	} else {
		result.Shoot = val
	}
	return result, nil
}

func SetGardenerClusterSummary(gc *unstructured.Unstructured, summary GardenerClusterSummary) error {
	spec := map[string]interface{}{
		"kubeconfig": map[string]interface{}{
			"secret": map[string]interface{}{
				"name":      summary.Name,
				"key":       summary.Key,
				"namespace": summary.Namespace,
			},
		},
		"shoot": map[string]interface{}{
			"name": summary.Shoot,
		},
	}
	return unstructured.SetNestedMap(gc.Object, spec, "spec")
}

type GardenerClusterSummary struct {
	Key       string
	Name      string
	Namespace string
	Shoot     string
}
