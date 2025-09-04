package clientspy

import (
	"context"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClientSpy interface {
	CreateCallCount() int64
	UpdateCallCount() int64
	DeleteCallCount() int64
	PatchCallCount() int64
	Client() client.WithWatch
	SetClient(client client.WithWatch)
}

type clientSpy struct {
	client          client.WithWatch
	createCallCount int64
	updateCallCount int64
	deleteCallCount int64
	patchCallCount  int64
}

func (clientSpy *clientSpy) Apply(ctx context.Context, obj runtime.ApplyConfiguration, opts ...client.ApplyOption) error {
	return clientSpy.client.Apply(ctx, obj, opts...)
}

func (s *clientSpy) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	s.createCallCount++
	return s.client.Create(ctx, obj, opts...)
}

func (s *clientSpy) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	s.deleteCallCount++
	return s.client.Delete(ctx, obj, opts...)
}

func (s *clientSpy) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	s.updateCallCount++
	return s.client.Update(ctx, obj, opts...)
}

func (s *clientSpy) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	s.patchCallCount++
	return s.client.Patch(ctx, obj, patch, opts...)
}

func (s *clientSpy) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return s.client.DeleteAllOf(ctx, obj, opts...)
}

func (s *clientSpy) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return s.client.Get(ctx, key, obj, opts...)
}

func (s *clientSpy) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return s.client.List(ctx, list, opts...)
}

func (s *clientSpy) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return s.client.GroupVersionKindFor(obj)
}

func (s *clientSpy) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return s.client.IsObjectNamespaced(obj)
}

func (s *clientSpy) RESTMapper() meta.RESTMapper {
	return s.client.RESTMapper()
}

func (s *clientSpy) Scheme() *runtime.Scheme {
	return s.client.Scheme()
}

func (s *clientSpy) Status() client.SubResourceWriter {
	return s.client.Status()
}

func (s *clientSpy) SubResource(subResource string) client.SubResourceClient {
	return s.client.SubResource(subResource)
}

func (s *clientSpy) Watch(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) (watch.Interface, error) {
	return s.client.Watch(ctx, obj, opts...)
}

func (s *clientSpy) CreateCallCount() int64 {
	return s.createCallCount
}

func (s *clientSpy) UpdateCallCount() int64 {
	return s.updateCallCount
}

func (s *clientSpy) DeleteCallCount() int64 {
	return s.deleteCallCount
}

func (s *clientSpy) PatchCallCount() int64 {
	return s.patchCallCount
}

func (s *clientSpy) Client() client.WithWatch {
	return s.client
}

func (s *clientSpy) SetClient(clnt client.WithWatch) {
	s.client = clnt
}

func NewClientSpy(clnt client.WithWatch) client.WithWatch {

	return &clientSpy{
		client:          clnt,
		createCallCount: 0,
		updateCallCount: 0,
		deleteCallCount: 0,
		patchCallCount:  0,
	}
}
