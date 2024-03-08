package composed

import (
	"context"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsMarkedForDeletion(obj client.Object) bool {
	if obj == nil {
		return false
	}
	val := reflect.ValueOf(obj)
	if val.IsNil() {
		return false
	}
	if obj.GetDeletionTimestamp() == nil {
		return false
	}
	if obj.GetDeletionTimestamp().IsZero() {
		return false
	}
	return true
}

func MarkedForDeletionPredicate(_ context.Context, state State) bool {
	return IsMarkedForDeletion(state.Obj())
}
