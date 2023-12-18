package composed

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"time"
)

func LoggerFromCtx(ctx context.Context) logr.Logger {
	return log.FromContext(ctx)
}

type State interface {
	Client() client.Client
	EventRecorder() record.EventRecorder
	Name() types.NamespacedName
	Obj() client.Object

	Result() ctrl.Result
	RequeueIfError(err error, msg ...string) error
	Stop(err error, msg ...string) error
	StopWithRequeue() error
	IsStopped() bool

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
		result:        ctrl.Result{},
	}
}

type baseState struct {
	result        ctrl.Result
	client        client.Client
	eventRecorder record.EventRecorder
	name          types.NamespacedName
	obj           client.Object
	stopped       bool
}

func (s *baseState) StopWithRequeue() error {
	s.stopped = true
	s.result.Requeue = true
	return nil
}

func (s *baseState) Stop(err error, msg ...string) error {
	s.stopped = true
	if err == nil {
		return nil
	}
	if len(msg) == 0 || len(msg[0]) == 0 {
		return err
	}
	return fmt.Errorf("%s: %w", msg[0], err)
}

func (s *baseState) IsStopped() bool {
	return s.stopped
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

func (s *baseState) Result() ctrl.Result {
	return s.result
}

func (s *baseState) RequeueIfError(err error, msg ...string) error {
	if err == nil {
		return nil
	}
	s.result.Requeue = true
	if len(msg) == 0 {
		return err
	}
	return fmt.Errorf("%s: %w", msg[0], err)
}

func (s *baseState) RequeueAfter(t time.Duration) {
	s.result.RequeueAfter = t
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
