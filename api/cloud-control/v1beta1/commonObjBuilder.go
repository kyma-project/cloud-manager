package v1beta1

import (
	"reflect"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// +kubebuilder:object:generate=false

type CommonObjBuilder[B any, T client.Object] struct {
	builder B
	Obj     T
}

func (b *CommonObjBuilder[B, T]) WithObj(obj T) B {
	b.Obj = obj
	return b.builder
}

func (b *CommonObjBuilder[B, T]) WithName(name string) B {
	b.Obj.SetName(name)
	return b.builder
}

func (b *CommonObjBuilder[B, T]) WithNamespace(namespace string) B {
	b.Obj.SetNamespace(namespace)
	return b.builder
}

func (b *CommonObjBuilder[B, T]) WithLabel(k, v string) B {
	if b.Obj.GetLabels() == nil {
		b.Obj.SetLabels(make(map[string]string))
	}
	b.Obj.GetLabels()[k] = v
	return b.builder
}

func (b *CommonObjBuilder[B, T]) WithLabels(labels map[string]string) B {
	if b.Obj.GetLabels() == nil {
		b.Obj.SetLabels(make(map[string]string))
	}
	for k, v := range labels {
		b.Obj.GetLabels()[k] = v
	}
	return b.builder
}

func (b *CommonObjBuilder[B, T]) WithAnnotation(k, v string) B {
	if b.Obj.GetAnnotations() == nil {
		b.Obj.SetAnnotations(make(map[string]string))
	}
	b.Obj.GetAnnotations()[k] = v
	return b.builder
}

func (b *CommonObjBuilder[B, T]) WithAnnotations(annotations map[string]string) B {
	if b.Obj.GetAnnotations() == nil {
		b.Obj.SetAnnotations(make(map[string]string))
	}
	for k, v := range annotations {
		b.Obj.GetAnnotations()[k] = v
	}
	return b.builder
}

func (b *CommonObjBuilder[B, T]) WithFinalizer(finalizer string) B {
	b.Obj.SetFinalizers(append(b.Obj.GetFinalizers(), finalizer))
	return b.builder
}

func (b *CommonObjBuilder[B, T]) Reset() B {
	t := reflect.TypeOf(b.Obj)
	b.Obj = reflect.New(t.Elem()).Interface().(T)
	return b.builder
}

func (b *CommonObjBuilder[B, T]) Build() T {
	return b.Obj
}
