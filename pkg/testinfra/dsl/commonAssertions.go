package dsl

import (
	"fmt"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func HavingDeletionTimestamp() ObjAssertion {
	return func(obj client.Object) error {
		if obj.GetDeletionTimestamp().IsZero() {
			return fmt.Errorf(
				"Expected object %T %s/%s to have deletion timestamp set, but it doesnt have it",
				obj,
				obj.GetNamespace(), obj.GetName(),
			)
		}
		return nil
	}
}
