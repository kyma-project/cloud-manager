package dsl

import (
	"context"
	"errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type ObjActionFunc func(obj client.Object)

type ObjAction interface {
	Apply(obj client.Object)
}

type ObjStatusAction interface {
	ObjAction
	ApplyOnStatus(obj client.Object)
}

type objAction struct {
	f ObjActionFunc
}

func (o *objAction) Apply(obj client.Object) {
	if o.f != nil {
		o.f(obj)
	}
}

type objStatusAction struct {
	f ObjActionFunc
}

func (o *objStatusAction) Apply(obj client.Object) {
	o.ApplyOnStatus(obj)
}

func (o *objStatusAction) ApplyOnStatus(obj client.Object) {
	if o.f != nil {
		o.f(obj)
	}
}

// =================

func NewObjActions(opts ...ObjAction) ObjActions {
	result := append(
		ObjActions{},
		opts...,
	)
	return result
}

type ObjActions []ObjAction

func (arr ObjActions) Append(actionsToAdd ...ObjAction) ObjActions {
	return append(arr, actionsToAdd...)
}

func (arr ObjActions) ApplyOnObject(obj client.Object) {
	for _, a := range arr {
		_, isStatusAction := a.(ObjStatusAction)
		if !isStatusAction {
			a.Apply(obj)
		}
	}
}

func (arr ObjActions) ApplyOnStatus(obj client.Object) {
	for _, a := range arr {
		sa, isStatusChanger := a.(ObjStatusAction)
		if isStatusChanger {
			sa.ApplyOnStatus(obj)
		}
	}
}

// ===========================================================

func WithNamespace(ns string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if obj.GetNamespace() == "" {
				obj.SetNamespace(ns)
			}
		},
	}
}

func WithName(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if obj.GetName() == "" {
				obj.SetName(name)
			}
		},
	}
}

func WithLabels(labels map[string]string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			if obj.GetLabels() == nil {
				obj.SetLabels(make(map[string]string))
			}
			for key, value := range labels {
				obj.GetLabels()[key] = value
			}
		},
	}
}

func RemoveFinalizer(name string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			controllerutil.RemoveFinalizer(obj, name)
		},
	}
}

func LoadAndUpdate(ctx context.Context, client client.Client, obj client.Object, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("the object for LoadAndUpdate() can not be nil")
	}

	err := LoadAndCheck(ctx, client, obj, NewObjActions())
	if err != nil {
		return err
	}

	NewObjActions(opts...).
		ApplyOnObject(obj)

	err = client.Update(ctx, obj)
	return err
}

func LoadAndDelete(ctx context.Context, client client.Client, obj client.Object) error {
	if obj == nil {
		return errors.New("the object for LoadAndDelete() can not be nil")
	}

	err := LoadAndCheck(ctx, client, obj, NewObjActions())
	if err != nil {
		return err
	}

	err = client.Delete(ctx, obj)
	return err
}

func IsDeleted(ctx context.Context, clnt client.Client, obj client.Object) bool {
	err := clnt.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	return err != nil && apierrors.IsNotFound(err)
}
