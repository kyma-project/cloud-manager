package v2

import (
	"context"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func IsMarkedForDeletion(obj client.Object) bool {
	if obj == nil {
		return true
	}
	val := reflect.ValueOf(obj)
	if val.IsNil() {
		return true
	}
	if obj.GetDeletionTimestamp() == nil {
		return false
	}
	if obj.GetDeletionTimestamp().IsZero() {
		return false
	}
	return true
}

var _ Predicate = MarkForDeletionPredicate

func MarkForDeletionPredicate(ctx context.Context) bool {
	state := StateFromCtx[State](ctx)
	return IsMarkedForDeletion(state.Obj())
}
