package dsl

import (
	"context"
	"errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func UpdateObj(ctx context.Context, clnt client.Client, obj client.Object, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("the object for UpdateObj() can not be nil")
	}

	NewObjActions(opts...).
		ApplyOnObject(obj)

	err := clnt.Update(ctx, obj)
	if err != nil {
		return err
	}

	return nil
}

func UpdateStatus(ctx context.Context, clnt client.Client, obj client.Object, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("the object for UpdateStatus() can not be nil")
	}

	NewObjActions(opts...).
		ApplyOnStatus(obj)

	err := clnt.Status().Update(ctx, obj)
	return err
}
