package dsl

import (
	"context"
	"errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CreateObj(ctx context.Context, clnt client.Client, obj client.Object, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("obj given to Create() can not be nil")
	}

	NewObjActions(opts...).
		ApplyOnObject(obj)

	err := clnt.Create(ctx, obj)

	return err
}
