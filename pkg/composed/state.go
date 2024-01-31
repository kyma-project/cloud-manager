package composed

import (
	"context"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

func NewStateClusterFromCluster(cluster cluster.Cluster) StateCluster {
	return NewStateCluster(cluster.GetClient(), cluster.GetEventRecorderFor("cloud-manager"), cluster.GetScheme())
}

func NewStateCluster(client client.Client,
	eventRecorder record.EventRecorder,
	scheme *runtime.Scheme,
) StateCluster {
	return &stateCluster{
		client:        client,
		eventRecorder: eventRecorder,
		scheme:        scheme,
	}
}

type StateCluster interface {
	K8sClient() client.Client
	EventRecorder() record.EventRecorder
	Scheme() *runtime.Scheme
}

type stateCluster struct {
	client        client.Client
	eventRecorder record.EventRecorder
	scheme        *runtime.Scheme
}

func (c *stateCluster) K8sClient() client.Client {
	return c.client
}

func (c *stateCluster) EventRecorder() record.EventRecorder {
	return c.eventRecorder
}

func (c *stateCluster) Scheme() *runtime.Scheme {
	return c.scheme
}

type State interface {
	// Cluster returns the used cluster
	Cluster() StateCluster

	// K8sClient returns the client to the cluster
	//
	// Deprecated: Use Cluster().K8sClient() instead
	K8sClient() client.Client

	// EventRecorder returns the event recorder of the connected cluster
	//
	// Deprecated: Use Cluster().EventRecorder() instead
	EventRecorder() record.EventRecorder

	// Scheme returns the Scheme
	//
	// Deprecated: Use Cluster().Scheme() instead
	Scheme() *runtime.Scheme

	Name() types.NamespacedName
	Obj() client.Object
	SetObj(client.Object)

	LoadObj(ctx context.Context, opts ...client.GetOption) error
	UpdateObj(ctx context.Context, opts ...client.UpdateOption) error
	UpdateObjStatus(ctx context.Context, opts ...client.SubResourceUpdateOption) error
}

type StateFactory interface {
	NewState(name types.NamespacedName, obj client.Object) State
}

func NewStateFactory(cluster StateCluster) StateFactory {
	return &stateFactory{cluster: cluster}
}

type stateFactory struct {
	cluster StateCluster
}

func (f *stateFactory) NewState(name types.NamespacedName, obj client.Object) State {
	return &baseState{
		cluster: f.cluster,
		name:    name,
		obj:     obj,
	}
}

// ========================================================================

type baseState struct {
	cluster StateCluster
	//client        client.Client
	//eventRecorder record.EventRecorder
	//scheme        *runtime.Scheme

	name types.NamespacedName
	obj  client.Object
}

func (s *baseState) Cluster() StateCluster {
	return s.cluster
}

func (s *baseState) K8sClient() client.Client {
	return s.cluster.K8sClient()
}

func (s *baseState) EventRecorder() record.EventRecorder {
	return s.cluster.EventRecorder()
}

func (s *baseState) Scheme() *runtime.Scheme {
	return s.cluster.Scheme()
}

func (s *baseState) Name() types.NamespacedName {
	return s.name
}

func (s *baseState) Obj() client.Object {
	return s.obj
}

func (s *baseState) SetObj(obj client.Object) {
	s.obj = obj
}

func (s *baseState) LoadObj(ctx context.Context, opts ...client.GetOption) error {
	return s.Cluster().K8sClient().Get(ctx, s.name, s.obj, opts...)
}

func (s *baseState) UpdateObj(ctx context.Context, opts ...client.UpdateOption) error {
	return s.Cluster().K8sClient().Update(ctx, s.Obj(), opts...)
}

func (s *baseState) UpdateObjStatus(ctx context.Context, opts ...client.SubResourceUpdateOption) error {
	return s.Cluster().K8sClient().Status().Update(ctx, s.Obj(), opts...)
}
