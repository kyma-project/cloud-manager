package composed

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func LoggerFromCtx(ctx context.Context) logr.Logger {
	return log.FromContext(ctx)
}

func LoggerIntoCtx(ctx context.Context, logger logr.Logger) context.Context {
	newCtx := log.IntoContext(ctx, logger)

	return newCtx
}

type State interface {
	Client() client.Client
	EventRecorder() record.EventRecorder
	Name() types.NamespacedName
	Obj() client.Object

	LoadObj(ctx context.Context, opts ...client.GetOption) error
	UpdateObj(ctx context.Context, opts ...client.UpdateOption) error
	UpdateObjStatus(ctx context.Context, opts ...client.SubResourceUpdateOption) error
}

func NewState(
	client client.Client,
	eventRecorder record.EventRecorder,
	name types.NamespacedName,
	obj client.Object,
) State {
	return &baseState{
		client:        client,
		eventRecorder: eventRecorder,
		name:          name,
		obj:           obj,
	}
}

type baseState struct {
	client        client.Client
	eventRecorder record.EventRecorder
	name          types.NamespacedName
	obj           client.Object
	nextCtxHanler func(ctx context.Context)
}

func (s *baseState) Client() client.Client {
	return s.client
}

func (s *baseState) EventRecorder() record.EventRecorder {
	return s.eventRecorder
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
