package dsl

import (
	"context"
	"errors"

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
	if o != nil {
		o.ApplyOnStatus(obj)
	}
}

func (o *objStatusAction) ApplyOnStatus(obj client.Object) {
	if o != nil && o.f != nil {
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

func (arr ObjActions) ApplyOnObject(obj client.Object) ObjActions {
	for _, a := range arr {
		_, isStatusAction := a.(ObjStatusAction)
		if !isStatusAction {
			a.Apply(obj)
		}
	}
	return arr
}

func (arr ObjActions) ApplyOnStatus(obj client.Object) ObjActions {
	for _, a := range arr {
		sa, isStatusChanger := a.(ObjStatusAction)
		if isStatusChanger {
			sa.ApplyOnStatus(obj)
		}
	}
	return arr
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
				if _, ok := obj.GetLabels()[key]; !ok {
					obj.GetLabels()[key] = value
				}
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

func AddFinalizer(finalizer string) ObjAction {
	return &objAction{
		f: func(obj client.Object) {
			controllerutil.AddFinalizer(obj, finalizer)
		},
	}
}

func Update(ctx context.Context, cl client.Client, obj client.Object, opts ...ObjAction) error {
	if obj == nil {
		return errors.New("the object for Update() can not be nil")
	}

	err := LoadAndCheck(ctx, cl, obj, NewObjActions())
	if err != nil {
		return err
	}

	NewObjActions(opts...).
		ApplyOnObject(obj)

	err = cl.Update(ctx, obj)
	return err
}

func Delete(ctx context.Context, clnt client.Client, obj client.Object) error {
	if obj == nil {
		return errors.New("the object for Delete() can not be nil")
	}

	err := LoadAndCheck(ctx, clnt, obj, NewObjActions())
	if err != nil {
		return err
	}

	err = clnt.Delete(ctx, obj)
	return err
}
