package util

import (
	"slices"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AllListsInGroup(group string, skipKinds []string, scheme *runtime.Scheme) []client.ObjectList {
	var result []client.ObjectList
	for gvk := range scheme.AllKnownTypes() {
		if gvk.Group != group {
			continue
		}
		if slices.Contains(skipKinds, gvk.Kind) {
			continue
		}
		if strings.HasSuffix(gvk.Kind, "List") {
			continue
		}
		listGvk := schema.GroupVersionKind{
			Group:   gvk.Group,
			Version: gvk.Version,
			Kind:    gvk.Kind + "List",
		}
		if !scheme.Recognizes(listGvk) {
			continue
		}
		listObj, err := scheme.New(listGvk)
		if runtime.IsNotRegisteredError(err) {
			continue
		}
		list, ok := listObj.(client.ObjectList)
		if !ok {
			continue
		}
		result = append(result, list)
	}
	return result
}
