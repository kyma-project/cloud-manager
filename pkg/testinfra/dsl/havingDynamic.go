package dsl

import (
	"errors"
	"fmt"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HavingFieldSet(fields ...string) ObjAssertion {
	return func(obj client.Object) error {
		if obj == nil {
			return errors.New("obj is nil")
		}
		u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return fmt.Errorf("field to conver obj %T %s to unstructured: %w", obj, client.ObjectKeyFromObject(obj), err)
		}
		val, found, err := unstructured.NestedFieldCopy(u, fields...)
		if err != nil {
			return fmt.Errorf("field to get value %v from unstructured: %w", fields, err)
		}
		if !found {
			return fmt.Errorf("path %v does not exist", fields)
		}
		v := reflect.ValueOf(val)
		if v.IsZero() {
			return fmt.Errorf("value at path %v is zero", fields)
		}
		return nil
	}
}

func HavingFieldValue(expectedValue any, fields ...string) ObjAssertion {
	return func(obj client.Object) error {
		if obj == nil {
			return errors.New("obj is nil")
		}
		u, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return fmt.Errorf("field to conver obj %T %s to unstructured: %w", obj, client.ObjectKeyFromObject(obj), err)
		}
		val, found, err := unstructured.NestedFieldCopy(u, fields...)
		if err != nil {
			return fmt.Errorf("field to get value %v from unstructured: %w", fields, err)
		}
		if !found {
			return fmt.Errorf("path %v does not exist", fields)
		}
		if reflect.DeepEqual(val, expectedValue) {
			return nil
		}
		return fmt.Errorf("value at path %v is %v, expectedValue %v", fields, val, expectedValue)
	}
}
