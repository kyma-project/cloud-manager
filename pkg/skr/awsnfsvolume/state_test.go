package awsnfsvolume

import (
	"context"

	composed "github.com/kyma-project/cloud-manager/pkg/composed"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testStateFactory(k8sClient client.WithWatch, obj client.Object) *State {

	cluster := composed.NewStateCluster(k8sClient, k8sClient, nil, k8sClient.Scheme())
	return &State{
		State:          composed.NewStateFactory(cluster).NewState(types.NamespacedName{}, obj),
		KcpCluster:     nil,
		SkrIpRange:     nil,
		KcpNfsInstance: nil,
		Volume:         nil,
		PVC:            nil,
	}
}

type ClientSpy interface {
	CreateCallCount() int64
	UpdateCallCount() int64
	DeleteCallCount() int64
	PatchCallCount() int64
}

type clientSpy struct {
	client          client.WithWatch
	createCallCount int64
	updateCallCount int64
	deleteCallCount int64
	patchCallCount  int64
}

// Create saves the object obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (clientSpy *clientSpy) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	clientSpy.createCallCount++
	return clientSpy.client.Create(ctx, obj, opts...)
}

// Delete deletes the given obj from Kubernetes cluster.
func (clientSpy *clientSpy) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	clientSpy.deleteCallCount++
	return clientSpy.Delete(ctx, obj, opts...)
}

// Update updates the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (clientSpy *clientSpy) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	clientSpy.updateCallCount++
	return clientSpy.client.Update(ctx, obj, opts...)
}

// Patch patches the given obj in the Kubernetes cluster. obj must be a
// struct pointer so that obj can be updated with the content returned by the Server.
func (clientSpy *clientSpy) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	clientSpy.patchCallCount++
	return clientSpy.client.Patch(ctx, obj, patch, opts...)
}

// DeleteAllOf deletes all objects of the given type matching the given options.
func (clientSpy *clientSpy) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return clientSpy.client.DeleteAllOf(ctx, obj, opts...)
}

// Get retrieves an obj for the given object key from the Kubernetes Cluster.
// obj must be a struct pointer so that obj can be updated with the response
// returned by the Server.
func (clientSpy *clientSpy) Get(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
	return clientSpy.client.Get(ctx, key, obj, opts...)
}

// List retrieves list of objects for a given namespace and list options. On a
// successful call, Items field in the list will be populated with the
// result returned from the server.
func (clientSpy *clientSpy) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	return clientSpy.client.List(ctx, list, opts...)
}

// GroupVersionKindFor implements client.WithWatch.
func (clientSpy *clientSpy) GroupVersionKindFor(obj runtime.Object) (schema.GroupVersionKind, error) {
	return clientSpy.client.GroupVersionKindFor(obj)
}

// IsObjectNamespaced implements client.WithWatch.
func (clientSpy *clientSpy) IsObjectNamespaced(obj runtime.Object) (bool, error) {
	return clientSpy.client.IsObjectNamespaced(obj)
}

// RESTMapper implements client.WithWatch.
func (clientSpy *clientSpy) RESTMapper() meta.RESTMapper {
	return clientSpy.client.RESTMapper()
}

// Scheme implements client.WithWatch.
func (clientSpy *clientSpy) Scheme() *runtime.Scheme {
	return clientSpy.client.Scheme()
}

// Status implements client.WithWatch.
func (clientSpy *clientSpy) Status() client.SubResourceWriter {
	return clientSpy.client.Status()
}

// SubResource implements client.WithWatch.
func (clientSpy *clientSpy) SubResource(subResource string) client.SubResourceClient {
	return clientSpy.client.SubResource(subResource)
}

// Watch implements client.WithWatch.
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

func NewClientSpy(scheme *runtime.Scheme) client.WithWatch {

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	return &clientSpy{
		client:          fakeClient,
		createCallCount: 0,
		updateCallCount: 0,
		deleteCallCount: 0,
		patchCallCount:  0,
	}
}
