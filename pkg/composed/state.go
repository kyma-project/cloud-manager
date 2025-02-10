package composed

import (
	"context"
	"fmt"
	"github.com/kyma-project/cloud-manager/pkg/common"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NewStateClusterFromCluster(cluster cluster.Cluster) StateCluster {
	return NewStateCluster(cluster.GetClient(), cluster.GetAPIReader(), cluster.GetEventRecorderFor("cloud-manager"), cluster.GetScheme())
}

func NewStateCluster(
	client client.Client,
	reader client.Reader,
	eventRecorder record.EventRecorder,
	scheme *runtime.Scheme,
) StateCluster {
	return &stateCluster{
		client:        client,
		reader:        reader,
		eventRecorder: eventRecorder,
		scheme:        scheme,
	}
}

type StateCluster interface {
	K8sClient() client.Client
	ApiReader() client.Reader
	EventRecorder() record.EventRecorder
	Scheme() *runtime.Scheme
}

type stateCluster struct {
	client        client.Client
	reader        client.Reader
	eventRecorder record.EventRecorder
	scheme        *runtime.Scheme
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
	PatchObjStatus(ctx context.Context) error

	PatchObjAddFinalizer(ctx context.Context, f string) (bool, error)
	PatchObjRemoveFinalizer(ctx context.Context, f string) (bool, error)
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

func (s *baseState) PatchObjStatus(ctx context.Context) error {
	return PatchObjStatus(ctx, s.Obj(), s.Cluster().K8sClient())
}

// PatchObjAddFinalizer uses controllerutil.AddFinalizer() to add finalizer, if it returns false
// meaning that object already had that finalizer and that object is not modified it returns nil,
// if the finalizer didn't exist and object is modified, then it
// patches obj with MergePatchType. Finalizer name f must consist of alphanumeric
// characters, '-', '_' or '.', and must start and end with an alphanumeric character (e.g. 'MyName',
// or 'my.name',  or '123-abc', regex used for validation is '([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]')"}
func (s *baseState) PatchObjAddFinalizer(ctx context.Context, f string) (bool, error) {
	return PatchObjAddFinalizer(ctx, f, s.Obj(), s.Cluster().K8sClient())
}

// PatchObjRemoveFinalizer patches obj with JSONPatchType removing the specified finalizer name
func (s *baseState) PatchObjRemoveFinalizer(ctx context.Context, f string) (bool, error) {
	return PatchObjRemoveFinalizer(ctx, f, s.Obj(), s.Cluster().K8sClient())
}

func PatchObjStatus(ctx context.Context, obj client.Object, clnt client.StatusClient) error {
	objToPatch := obj
	if objClonable, ok := obj.(ObjWithCloneForPatchStatus); ok {
		objToPatch = objClonable.CloneForPatchStatus()
	}
	return clnt.Status().Patch(ctx, objToPatch, client.Apply, client.ForceOwnership, client.FieldOwner(common.FieldOwner))
}

func PatchObjAddFinalizer(ctx context.Context, f string, obj client.Object, clnt client.Writer) (bool, error) {
	added := controllerutil.AddFinalizer(obj, f)
	if !added {
		return false, nil
	}
	p := []byte(fmt.Sprintf(`{"metadata": {"finalizers":["%s"]}}`, f))
	return true, clnt.Patch(ctx, obj, client.RawPatch(types.MergePatchType, p))
}

func PatchObjMergeAnnotation(ctx context.Context, k, v string, obj client.Object, clnt client.Writer) (bool, error) {
	if obj.GetAnnotations() != nil && obj.GetAnnotations()[k] == v {
		return false, nil
	}
	if obj.GetAnnotations() == nil {
		obj.SetAnnotations(map[string]string{})
	}
	obj.GetAnnotations()[k] = v
	p := []byte(fmt.Sprintf(`{"metadata": {"annotations":{"%s": "%s"}}}`, k, v))
	return true, clnt.Patch(ctx, obj, client.RawPatch(types.MergePatchType, p))
}

func PatchObjRemoveFinalizer(ctx context.Context, f string, obj client.Object, clnt client.Writer) (bool, error) {
	idx := -1
	for i, s := range obj.GetFinalizers() {
		if s == f {
			idx = i
			break
		}
	}
	if idx == -1 {
		return false, nil
	}
	controllerutil.RemoveFinalizer(obj, f)
	p := []byte(fmt.Sprintf(`[{"op": "remove", "path": "/metadata/finalizers/%d"}]`, idx))
	return true, clnt.Patch(ctx, obj, client.RawPatch(types.JSONPatchType, p))
}
