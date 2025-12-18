package dsl

import (
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HavingLabels(expectedLabels map[string]string) ObjAssertion {
	return func(obj client.Object) error {
		actualLabels := obj.GetLabels()
		if actualLabels == nil {
			actualLabels = make(map[string]string)
		}

		for key, expectedValue := range expectedLabels {
			actualValue, exists := actualLabels[key]
			if !exists {
				return fmt.Errorf(
					"expected object %T %s/%s to have label %s, but it doesn't exist",
					obj,
					obj.GetNamespace(), obj.GetName(),
					key,
				)
			}
			if actualValue != expectedValue {
				return fmt.Errorf(
					"expected object %T %s/%s to have label %s=%s, but got %s=%s",
					obj,
					obj.GetNamespace(), obj.GetName(),
					key, expectedValue,
					key, actualValue,
				)
			}
		}
		return nil
	}
}

func HavingLabel(key, expectedValue string) ObjAssertion {
	return HavingLabels(map[string]string{key: expectedValue})
}

func HavingLabelKeys(keys ...string) ObjAssertion {
	return func(obj client.Object) error {
		actualLabels := obj.GetLabels()
		if actualLabels == nil {
			actualLabels = make(map[string]string)
		}

		for _, key := range keys {
			if _, exists := actualLabels[key]; !exists {
				return fmt.Errorf(
					"expected object %T %s/%s to have label %s, but it doesn't exist",
					obj,
					obj.GetNamespace(), obj.GetName(),
					key,
				)
			}
		}
		return nil
	}
}

func NotHavingLabels(labelsToCheck ...string) ObjAssertion {
	return func(obj client.Object) error {
		actualLabels := obj.GetLabels()
		if actualLabels == nil {
			return nil
		}

		for _, key := range labelsToCheck {
			if _, exists := actualLabels[key]; exists {
				return fmt.Errorf(
					"expected object %T %s/%s to NOT have label %s, but it exists with value %s",
					obj,
					obj.GetNamespace(), obj.GetName(),
					key, actualLabels[key],
				)
			}
		}
		return nil
	}
}

func NotHavingLabel(key string) ObjAssertion {
	return NotHavingLabels(key)
}

func HavingAnnotations(expectedAnnotations map[string]string) ObjAssertion {
	return func(obj client.Object) error {
		actualAnnotations := obj.GetAnnotations()
		if actualAnnotations == nil {
			actualAnnotations = make(map[string]string)
		}

		for key, expectedValue := range expectedAnnotations {
			actualValue, exists := actualAnnotations[key]
			if !exists {
				return fmt.Errorf(
					"expected object %T %s/%s to have annotation %s, but it doesn't exist",
					obj,
					obj.GetNamespace(), obj.GetName(),
					key,
				)
			}
			if actualValue != expectedValue {
				return fmt.Errorf(
					"expected object %T %s/%s to have annotation %s=%s, but got %s=%s",
					obj,
					obj.GetNamespace(), obj.GetName(),
					key, expectedValue,
					key, actualValue,
				)
			}
		}
		return nil
	}
}

func HavingAnnotation(key, expectedValue string) ObjAssertion {
	return HavingAnnotations(map[string]string{key: expectedValue})
}

func NotHavingAnnotations(annotationsToCheck ...string) ObjAssertion {
	return func(obj client.Object) error {
		actualAnnotations := obj.GetAnnotations()
		if actualAnnotations == nil {
			return nil
		}

		for _, key := range annotationsToCheck {
			if _, exists := actualAnnotations[key]; exists {
				return fmt.Errorf(
					"expected object %T %s/%s to NOT have annotation %s, but it exists with value %s",
					obj,
					obj.GetNamespace(), obj.GetName(),
					key, actualAnnotations[key],
				)
			}
		}
		return nil
	}
}

func NotHavingAnnotation(key string) ObjAssertion {
	return NotHavingAnnotations(key)
}
