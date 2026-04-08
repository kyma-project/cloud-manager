package dsl

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/elliotchance/pie/v2"
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
			return fmt.Errorf("field to get value %s from unstructured: %w", pie.Join(fields, "."), err)
		}
		if !found {
			return fmt.Errorf("path %s does not exist", pie.Join(fields, "."))
		}
		v := reflect.ValueOf(val)
		// check if zero - aka value type with default value
		if x, ok := v.Interface().(isZeroer); ok && x.IsZero() {
			return fmt.Errorf("value at path %s is zero", pie.Join(fields, "."))
		}
		if v.IsZero() {
			return fmt.Errorf("value at path %s is zero", pie.Join(fields, "."))
		}
		return nil
	}
}

// HavingFieldValue returns ObjAssertion that checks if value at specified property path
// equals to the given expectedValue. It converts the client.Object to unstructured, gets the nested field
// at given path and uses reflect.DeepEqual to compare it with expectedValue.
// If the field does not exist, or the value is not equal to expectedValue, it returns an error.
// NOTE!!!: some types when converted to unstructured are marshaled into string, for example
// resource.Quantity, metav1.Time... The func does not pay attention to those details, and expects
// that caller to know what type it will be when converted to unstructured. So if you want to compare
// resource.Quantity, you should pass expectedValue as string, for example
// HavingFieldValue("1G", "status", "capacity") instead of HavingFieldValue(resource.MustParse("1G"), "status", "capacity")
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
			return fmt.Errorf("field to get value %s from unstructured: %w", pie.Join(fields, "."), err)
		}
		if !found {
			return fmt.Errorf("path %s does not exist", pie.Join(fields, "."))
		}

		if reflect.DeepEqual(val, expectedValue) {
			return nil
		}
		return fmt.Errorf("value at path %s is '%v', expectedValue '%v'", pie.Join(fields, "."), val, expectedValue)
	}
}

type isZeroer interface {
	IsZero() bool
}
