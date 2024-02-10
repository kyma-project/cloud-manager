package dsl

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
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
