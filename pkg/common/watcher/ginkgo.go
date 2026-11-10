package watcher

import (
	ginkgo "github.com/onsi/ginkgo/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func NewGinkgoTestStart() client.Object {
	spec := ginkgo.CurrentSpecReport()
	return &unstructured.Unstructured{Object: map[string]any{
		"kind": "GinkgoTestStart",
		"spec": map[string]any{
			"id":           spec.LeafNodeLocation.String(),
			"nodeTitles":   spec.ContainerHierarchyTexts,
			"leafTitle":    spec.LeafNodeText,
			"fileLocation": spec.LeafNodeLocation.String(),
		},
	}}
}

func NewGinkgoTestEnd() client.Object {
	spec := ginkgo.CurrentSpecReport()
	return &unstructured.Unstructured{Object: map[string]any{
		"kind": "GinkgoTestEnd",
		"spec": map[string]any{
			"id":           spec.LeafNodeLocation.String(),
			"nodeTitles":   spec.ContainerHierarchyTexts,
			"leafTitle":    spec.LeafNodeText,
			"fileLocation": spec.LeafNodeLocation.String(),
		},
	}}
}
