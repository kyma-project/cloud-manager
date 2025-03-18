package v2

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
)

const (
	DefaultClusterID = "default"
)

func NewDefaultStateClusterFromCluster(cluster cluster.Cluster) StateCluster {
	return NewStateCluster(DefaultClusterID, cluster.GetClient(), cluster.GetAPIReader(), cluster.GetEventRecorderFor("gpu-driver"), cluster.GetScheme())
}

func NewStateClusterFromCluster(id string, cluster cluster.Cluster) StateCluster {
	return NewStateCluster(id, cluster.GetClient(), cluster.GetAPIReader(), cluster.GetEventRecorderFor("gpu-driver"), cluster.GetScheme())
}

func NewStateCluster(
	clusterId string,
	client client.Client,
	reader client.Reader,
	eventRecorder record.EventRecorder,
	scheme *runtime.Scheme,
) StateCluster {
	return &stateCluster{
		clusterId:     clusterId,
		client:        client,
		reader:        reader,
		eventRecorder: eventRecorder,
		scheme:        scheme,
	}
}

type StateCluster interface {
	ClusterID() string
	K8sClient() client.Client
	ApiReader() client.Reader
	EventRecorder() record.EventRecorder
	Scheme() *runtime.Scheme
}

type stateCluster struct {
	clusterId     string
	client        client.Client
	reader        client.Reader
	eventRecorder record.EventRecorder
	scheme        *runtime.Scheme
}

func (c *stateCluster) ClusterID() string {
	return c.clusterId
}

func (c *stateCluster) K8sClient() client.Client {
	return c.client
}

func (c *stateCluster) ApiReader() client.Reader {
	return c.reader
}

func (c *stateCluster) EventRecorder() record.EventRecorder {
	return c.eventRecorder
}

func (c *stateCluster) Scheme() *runtime.Scheme {
	return c.scheme
}

type ctxClusterKey string

func ClusterToCtx(ctx context.Context, cluster StateCluster) context.Context {
	key := ctxClusterKey(fmt.Sprintf("cluster_%s", cluster.ClusterID()))
	return context.WithValue(ctx, key, cluster)
}

func ClusterFromCtx(ctx context.Context, clusterID string) StateCluster {
	key := ctxClusterKey(fmt.Sprintf("cluster_%s", clusterID))
	x := ctx.Value(key)
	return x.(StateCluster)
}

func DefaultClusterFromCtx(ctx context.Context) StateCluster {
	return ClusterFromCtx(ctx, DefaultClusterID)
}

// State ===================

type State interface {
	Name() types.NamespacedName
	Obj() client.Object
	SetObj(obj client.Object)
}

type baseState struct {
	name types.NamespacedName
	obj  client.Object
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

// ===============================

func InitState(ctx context.Context, name types.NamespacedName, obj client.Object) context.Context {
	return StateToCtx(ctx, &baseState{
		name: name,
		obj:  obj,
	})
}

type ctxStateKeyType string

const ctxStateKey = ctxStateKeyType("state")

func StateToCtx(ctx context.Context, state any) context.Context {
	return context.WithValue(ctx, ctxStateKey, state)
}

func StateFromCtx[T State](ctx context.Context) T {
	var none T
	x := ctx.Value(ctxStateKey)
	r, ok := x.(T)
	if !ok {
		return none
	}
	return r
}
