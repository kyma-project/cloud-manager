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

func (clientSpy *clientSpy) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	clientSpy.createCallCount++
	return clientSpy.client.Create(ctx, obj, opts...)
}

func (clientSpy *clientSpy) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	clientSpy.deleteCallCount++
	return clientSpy.client.Delete(ctx, obj, opts...)
}

func (clientSpy *clientSpy) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	clientSpy.updateCallCount++
	return clientSpy.client.Update(ctx, obj, opts...)
}

func (clientSpy *clientSpy) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	clientSpy.patchCallCount++
	return clientSpy.client.Patch(ctx, obj, patch, opts...)
}

func (clientSpy *clientSpy) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return clientSpy.client.DeleteAllOf(ctx, obj, opts...)
}

func (clientSpy *clientSpy) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return clientSpy.client.Get(ctx, key, obj, opts...)
}

func (clientSpy *clientSpy) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return clientSpy.client.List(ctx, list, opts...)
}

func (clientSpy *clientSpy) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return clientSpy.client.GroupVersionKindFor(obj)
}

func (clientSpy *clientSpy) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return clientSpy.client.IsObjectNamespaced(obj)
}

func (clientSpy *clientSpy) RESTMapper() meta.RESTMapper {
	return clientSpy.client.RESTMapper()
}

func (clientSpy *clientSpy) Scheme() *runtime.Scheme {
	return clientSpy.client.Scheme()
}

func (clientSpy *clientSpy) Status() client.SubResourceWriter {
	return clientSpy.client.Status()
}

func (clientSpy *clientSpy) SubResource(subResource string) client.SubResourceClient {
	return clientSpy.client.SubResource(subResource)
}

func (clientSpy *clientSpy) Watch(ctx context.Context, obj client.ObjectList, opts ...client.ListOption) (watch.Interface, error) {
	return clientSpy.client.Watch(ctx, obj, opts...)
}

func (clientSpy *clientSpy) CreateCallCount() int64 {
	return clientSpy.createCallCount
}

func (clientSpy *clientSpy) UpdateCallCount() int64 {
	return clientSpy.updateCallCount
}

func (clientSpy *clientSpy) DeleteCallCount() int64 {
	return clientSpy.deleteCallCount
}

func (clientSpy *clientSpy) PatchCallCount() int64 {
	return clientSpy.patchCallCount
}

func (clientSpy *clientSpy) Client() client.WithWatch {
	return clientSpy.client
}

func (clientSpy *clientSpy) SetClient(cl client.WithWatch) {
	clientSpy.client = cl
}

func NewClientSpy(client client.WithWatch) client.WithWatch {

	return &clientSpy{
		client:          client,
		createCallCount: 0,
		updateCallCount: 0,
		deleteCallCount: 0,
		patchCallCount:  0,
	}
}
