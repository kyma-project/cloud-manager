package composed

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type State interface {
	Client() client.Client
	EventRecorder() record.EventRecorder
	Scheme() *runtime.Scheme
	Name() types.NamespacedName
	Obj() client.Object

	LoadObj(ctx context.Context, opts ...client.GetOption) error
	UpdateObj(ctx context.Context, opts ...client.UpdateOption) error
	UpdateObjStatus(ctx context.Context, opts ...client.SubResourceUpdateOption) error
}

type StateFactory interface {
	NewState(name types.NamespacedName, obj client.Object) State
}

func NewStateFactory(
	client client.Client,
	eventRecorder record.EventRecorder,
	scheme *runtime.Scheme,
) StateFactory {
	return &stateFactory{
		client:        client,
		eventRecorder: eventRecorder,
		scheme:        scheme,
	}
}

type stateFactory struct {
	client        client.Client
	eventRecorder record.EventRecorder
	scheme        *runtime.Scheme
}

func (f *stateFactory) NewState(name types.NamespacedName, obj client.Object) State {
	return &baseState{
		client:        f.client,
		eventRecorder: f.eventRecorder,
		scheme:        f.scheme,
		name:          name,
		obj:           obj,
	}
}

// ========================================================================

type baseState struct {
	client        client.Client
	eventRecorder record.EventRecorder
	scheme        *runtime.Scheme
	name          types.NamespacedName
	obj           client.Object
}

func (s *baseState) Client() client.Client {
	return s.client
}

func (s *baseState) EventRecorder() record.EventRecorder {
	return s.eventRecorder
}

func (s *baseState) Scheme() *runtime.Scheme {
	return s.scheme
}

func (s *baseState) Name() types.NamespacedName {
	return s.name
}

func (s *baseState) Obj() client.Object {
	return s.obj
}

func (s *baseState) LoadObj(ctx context.Context, opts ...client.GetOption) error {
	return s.client.Get(ctx, s.name, s.obj, opts...)
}

func (s *baseState) UpdateObj(ctx context.Context, opts ...client.UpdateOption) error {
	return s.client.Update(ctx, s.Obj(), opts...)
}

func (s *baseState) UpdateObjStatus(ctx context.Context, opts ...client.SubResourceUpdateOption) error {
	return s.Client().Status().Update(ctx, s.Obj(), opts...)
}
