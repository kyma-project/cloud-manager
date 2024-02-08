package dsl

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AssertHasLabelValue(name, expected string) ObjAssertion {
	return func(obj client.Object) error {
		if obj.GetLabels() == nil {
			return fmt.Errorf(
				"expected object %T %s/%s to have label %s: %s, but it has no labels",
				obj,
				obj.GetNamespace(), obj.GetName(),
				name, expected,
			)
		}

		actual, exists := obj.GetLabels()[name]
		if !exists {
			return fmt.Errorf(
				"expected object %T %s/%s to have label %s: %s, but it's not set",
				obj,
				obj.GetNamespace(), obj.GetName(),
				name, expected,
			)
		}

		if actual != expected {
			return fmt.Errorf(
				"expected object %T %s/%s to have label %s: %s, but it's value is %s",
				obj,
				obj.GetNamespace(), obj.GetName(),
				name, expected,
				actual,
			)
		}
		return nil
	}
}
