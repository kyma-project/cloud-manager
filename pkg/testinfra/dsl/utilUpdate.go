package dsl

import "sigs.k8s.io/controller-runtime/pkg/client"

func copyForUpdate(existing, desired client.Object) {
	// labels
	if desired.GetLabels() != nil {
		if existing.GetLabels() == nil {
			existing.SetLabels(map[string]string{})
		}
		for k, v := range desired.GetLabels() {
			existing.GetLabels()[k] = v
		}
	}

	// annotations
	if desired.GetAnnotations() != nil {
		if existing.GetAnnotations() == nil {
			existing.SetAnnotations(map[string]string{})
		}
		for k, v := range desired.GetAnnotations() {
			existing.GetAnnotations()[k] = v
		}
	}
}
